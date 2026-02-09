# GORM Gen

Friendly & Safer GORM powered by Code Generation.

[![Release](https://img.shields.io/github/v/release/go-gorm/gen)](https://github.com/go-gorm/gen/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-gorm/gen)](https://goreportcard.com/report/github.com/go-gorm/gen)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![OpenIssue](https://img.shields.io/github/issues/go-gorm/gen)](https://github.com/go-gorm/gen/issues?q=is%3Aopen+is%3Aissue)
[![ClosedIssue](https://img.shields.io/github/issues-closed/go-gorm/gen)](https://github.com/go-gorm/gen/issues?q=is%3Aissue+is%3Aclosed)
[![TODOs](https://badgen.net/https/api.tickgit.com/badgen/github.com/go-gorm/gen)](https://www.tickgit.com/browse?repo=github.com/go-gorm/gen)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/gorm.io/gen?tab=doc)

## Overview

- Idiomatic & Reusable API from Dynamic Raw SQL
- 100% Type-safe DAO API without `interface{}`
- Database To Struct follows GORM conventions
- GORM under the hood, supports all features, plugins, DBMS that GORM supports

## Getting Started

* Gen Guides [https://gorm.io/gen/index.html](https://gorm.io/gen/index.html)
* GORM Guides [http://gorm.io/docs](http://gorm.io/docs)

## Usage

Create a generator with `gen.Config`, then generate models and query code:

```go
package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	db, _ := gorm.Open(mysql.Open("user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True"), &gorm.Config{})

	cfg := gen.Config{
		OutPath:      "./dal/query",
		ModelPkgPath: "./dal/model",
		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
	}

	g := gen.NewGenerator(cfg)
	g.UseDB(db)
	g.GenerateAllTable()
	g.Execute()
}
```

## Soft delete

By default, a column named `deleted_at` with type `time.Time` is generated as `gorm.DeletedAt`. For custom columns use:

| Need | Config |
|------|--------|
| Default `deleted_at` | Don’t call any soft-delete option |
| Flag only (e.g. `is_delete` 0/1) | `cfg.WithSoftDeleteFlag("is_delete")` |
| Time only (e.g. `update_time` as delete time) | `cfg.WithSoftDeleteAt("update_time")` |
| Flag + time (mixed) | `cfg.WithSoftDeleteFlag("is_delete")` and `cfg.WithSoftDeleteAt("update_time")` |

Example:

```go
cfg := gen.Config{OutPath: "./query", ModelPkgPath: "./model"}
cfg.WithSoftDeleteFlag("is_delete")
cfg.WithSoftDeleteAt("update_time")

g := gen.NewGenerator(cfg)
g.UseDB(db)
g.GenerateAllTable()
g.Execute()
```

See [GORM Soft Delete](https://gorm.io/docs/delete.html#Soft-Delete) and `gorm.io/plugin/soft_delete` for mixed mode.

## Maintainers

[@riverchu](https://github.com/riverchu) [@iDer](https://github.com/idersec) [@qqxhb](https://github.com/qqxhb) [@dino-ma](https://github.com/dino-ma)

[@jinzhu](https://github.com/jinzhu)

## Contributing

[You can help to deliver a better GORM/Gen, check out things you can do](https://gorm.io/contribute.html)

## License

Released under the [MIT License](https://github.com/go-gorm/gen/blob/master/License)
