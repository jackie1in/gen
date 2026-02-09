package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"gorm.io/gen/internal/model"
)

// GenerateMode generate mode
type GenerateMode uint

const (
	// WithDefaultQuery create default query in generated code
	WithDefaultQuery GenerateMode = 1 << iota

	// WithoutContext generate code without context constrain
	WithoutContext

	// WithQueryInterface generate code with exported interface object
	WithQueryInterface

	// WithGeneric generate code with generic
	WithGeneric
)

// Config generator's basic configuration
type Config struct {
	db *gorm.DB // db connection

	OutPath      string // query code path
	OutFile      string // query code file name, default: gen.go
	ModelPkgPath string // generated model code's package name
	WithUnitTest bool   // generate unit test for query code

	// generate model global configuration
	FieldNullable     bool // generate pointer when field is nullable
	FieldCoverable    bool // generate pointer when field has default value, to fix problem zero value cannot be assign: https://gorm.io/docs/create.html#Default-Values
	FieldSignable     bool // detect integer field's unsigned type, adjust generated data type
	FieldWithIndexTag bool // generate with gorm index tag
	FieldWithTypeTag  bool // generate with gorm column type tag

	// Soft delete: use 0/1 flag column and a separate time column (GORM mixed mode).
	// e.g. WithSoftDeleteFlag("is_delete") and WithSoftDeleteAt("update_time") for mixed mode.
	SoftDeleteFlagColumn string // column name for delete flag (0=active, 1=deleted), e.g. "is_delete"
	SoftDeleteTimeColumn string // column name for delete time, e.g. "update_time"

	Mode GenerateMode // generate mode

	queryPkgName   string // generated query code's package name
	modelPkgPath   string // model pkg path in target project
	dbNameOpts     []model.SchemaNameOpt
	importPkgPaths []string

	// name strategy for syncing table from db
	tableNameNS func(tableName string) (targetTableName string)
	modelNameNS func(tableName string) (modelName string)
	fileNameNS  func(tableName string) (fileName string)

	dataTypeMap    map[string]func(columnType gorm.ColumnType) (dataType string)
	fieldJSONTagNS func(columnName string) (tagContent string)

	modelOpts []ModelOpt
}

// WithOpts set global  model options
func (cfg *Config) WithOpts(opts ...ModelOpt) {
	if cfg.modelOpts == nil {
		cfg.modelOpts = opts
	} else {
		cfg.modelOpts = append(cfg.modelOpts, opts...)
	}
}

// WithDbNameOpts set get database name function
func (cfg *Config) WithDbNameOpts(opts ...model.SchemaNameOpt) {
	if cfg.dbNameOpts == nil {
		cfg.dbNameOpts = opts
	} else {
		cfg.dbNameOpts = append(cfg.dbNameOpts, opts...)
	}
}

// WithTableNameStrategy specify table name naming strategy, only work when syncing table from db
func (cfg *Config) WithTableNameStrategy(ns func(tableName string) (targetTableName string)) {
	cfg.tableNameNS = ns
}

// WithModelNameStrategy specify model struct name naming strategy, only work when syncing table from db
func (cfg *Config) WithModelNameStrategy(ns func(tableName string) (modelName string)) {
	cfg.modelNameNS = ns
}

// WithFileNameStrategy specify file name naming strategy, only work when syncing table from db
func (cfg *Config) WithFileNameStrategy(ns func(tableName string) (fileName string)) {
	cfg.fileNameNS = ns
}

// WithDataTypeMap specify data type mapping relationship, only work when syncing table from db
func (cfg *Config) WithDataTypeMap(newMap map[string]func(columnType gorm.ColumnType) (dataType string)) {
	cfg.dataTypeMap = newMap
}

// WithJSONTagNameStrategy specify json tag naming strategy
func (cfg *Config) WithJSONTagNameStrategy(ns func(columnName string) (tagContent string)) {
	cfg.fieldJSONTagNS = ns
}

// WithSoftDeleteFlag sets the column name for the delete flag (0 = active, 1 = deleted).
// E.g. WithSoftDeleteFlag("is_delete"). Uses gorm.io/plugin/soft_delete when non-empty.
// Combine with WithSoftDeleteAt for mixed mode (flag + delete time).
func (cfg *Config) WithSoftDeleteFlag(column string) {
	cfg.SoftDeleteFlagColumn = strings.TrimSpace(column)
}

// WithSoftDeleteAt sets the column name for the delete time (e.g. "deleted_at", "update_time").
// When set, that column with type time.Time is generated as gorm.DeletedAt.
// Use alone for time-only soft delete, or with WithSoftDeleteFlag for mixed mode (flag + time).
func (cfg *Config) WithSoftDeleteAt(column string) {
	cfg.SoftDeleteTimeColumn = strings.TrimSpace(column)
}

func (cfg *Config) CheckSoftDeletePlugin() {
	if cfg.SoftDeleteFlagColumn != "" || cfg.SoftDeleteTimeColumn != "" {
		cfg.WithImportPkgPath("gorm.io/plugin/soft_delete")
	}
}

// WithImportPkgPath specify import package path
func (cfg *Config) WithImportPkgPath(paths ...string) {
	for i, path := range paths {
		path = strings.TrimSpace(path)
		if len(path) > 0 && path[0] != '"' && path[len(path)-1] != '"' { // without quote
			path = `"` + path + `"`
		}
		paths[i] = path
	}
	cfg.importPkgPaths = append(cfg.importPkgPaths, paths...)
}

// WithDataTypesNullType configures the types of fields to use their datatypes nullable counterparts.
/**
 *
 * @param {boolean} all - If true, all basic types of fields will be replaced with their `datatypes.Null[T]` types.
 *                        If false, only fields that are allowed to be null will be replaced with `datatypes.Null[T]` types.
 *
 * Examples:
 *
 * When `all` is true:
 * - `int64` will be replaced with `datatypes.NullInt64`
 * - `string` will be replaced with `datatypes.NullString`
 *
 * When `all` is false:
 * - Only fields that can be null (e.g., `*string` or `*int`) will be replaced with `datatypes.Null[T]` types.
 *
 * Note:
 * Ensure that proper error handling is implemented when converting
 * fields to their `datatypes.Null[T]` types to avoid runtime issues.
 */
func (cfg *Config) WithDataTypesNullType(all bool) {
	cfg.WithOpts(WithDataTypesNullType(all))
}

// Revise format path and db
func (cfg *Config) Revise() (err error) {
	if strings.TrimSpace(cfg.ModelPkgPath) == "" {
		cfg.ModelPkgPath = model.DefaultModelPkg
	}

	cfg.OutPath, err = filepath.Abs(cfg.OutPath)
	if err != nil {
		return fmt.Errorf("outpath is invalid: %w", err)
	}
	if cfg.OutPath == "" {
		cfg.OutPath = fmt.Sprintf(".%squery%s", string(os.PathSeparator), string(os.PathSeparator))
	}
	if cfg.OutFile == "" {
		cfg.OutFile = filepath.Join(cfg.OutPath, "gen.go")
	} else if !strings.Contains(cfg.OutFile, string(os.PathSeparator)) {
		cfg.OutFile = filepath.Join(cfg.OutPath, cfg.OutFile)
	}
	cfg.queryPkgName = filepath.Base(cfg.OutPath)

	if cfg.db == nil {
		cfg.db, _ = gorm.Open(tests.DummyDialector{})
	}

	return nil
}

func (cfg *Config) judgeMode(mode GenerateMode) bool { return cfg.Mode&mode != 0 }
