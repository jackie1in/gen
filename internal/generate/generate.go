package generate

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/jackie1in/gen/internal/model"
)

/*
** The feature of mapping table from database server to Golang struct
** Provided by @qqxhb
 */

func getFields(db *gorm.DB, conf *model.Config, columns []*model.Column) (fields []*model.Field) {
	for _, col := range columns {
		col.SetDataTypeMap(conf.DataTypeMap)
		col.WithNS(conf.FieldJSONTagNS)

		m := col.ToField(conf.FieldNullable, conf.FieldCoverable, conf.FieldSignable, conf.SoftDeleteTimeColumn)

		if filterField(m, conf.FilterOpts) == nil {
			continue
		}
		if _, ok := col.ColumnType.ColumnType(); ok && !conf.FieldWithTypeTag { // remove type tag if FieldWithTypeTag == false
			m.GORMTag.Remove("type")
		}

		// Custom soft delete (mixed mode): flag column 0/1 + time column
		if conf.SoftDeleteFlagColumn != "" && col.Name() == conf.SoftDeleteFlagColumn {
			m.Type = "soft_delete.DeletedAt"
			m.GORMTag.Set("softDelete", "flag")
			if conf.SoftDeleteTimeColumn != "" {
				timeFieldStructName := schemaName(db, conf.SoftDeleteTimeColumn)
				m.GORMTag.Set("DeletedAtField", timeFieldStructName)
			}
		}

		m = modifyField(m, conf.ModifyOpts)
		if ns, ok := db.NamingStrategy.(schema.NamingStrategy); ok {
			ns.SingularTable = true
			m.Name = ns.SchemaName(ns.TablePrefix + m.Name)
		} else if db.NamingStrategy != nil {
			m.Name = db.NamingStrategy.SchemaName(m.Name)
		}

		fields = append(fields, m)
	}
	for _, create := range conf.CreateOpts {
		m := create.Operator()(nil)
		if m.Relation != nil {
			if m.Relation.Model() != nil {
				stmt := gorm.Statement{DB: db}
				_ = stmt.Parse(m.Relation.Model())
				if stmt.Schema != nil {
					m.Relation.AppendChildRelation(ParseStructRelationShip(&stmt.Schema.Relationships)...)
				}
			}
			m.Type = strings.ReplaceAll(m.Type, conf.ModelPkg+".", "") // remove modelPkg in field's Type, avoid import error
		}

		fields = append(fields, m)
	}
	return fields
}

func filterField(m *model.Field, opts []model.FieldOption) *model.Field {
	for _, opt := range opts {
		if opt.Operator()(m) == nil {
			return nil
		}
	}
	return m
}

func modifyField(m *model.Field, opts []model.FieldOption) *model.Field {
	for _, opt := range opts {
		m = opt.Operator()(m)
	}
	return m
}

// schemaName returns the struct field name for a column (e.g. update_time -> UpdateTime)
func schemaName(db *gorm.DB, columnName string) string {
	if db == nil || db.NamingStrategy == nil {
		return columnName
	}
	if ns, ok := db.NamingStrategy.(schema.NamingStrategy); ok {
		ns.SingularTable = true
		return ns.SchemaName(columnName)
	}
	return db.NamingStrategy.SchemaName(columnName)
}

// get mysql db' name
var modelNameReg = regexp.MustCompile(`^\w+$`)

func checkStructName(name string) error {
	if name == "" {
		return nil
	}
	if !modelNameReg.MatchString(name) {
		return fmt.Errorf("model name cannot contains invalid character")
	}
	if name[0] < 'A' || name[0] > 'Z' {
		return fmt.Errorf("model name must be initial capital")
	}
	return nil
}
