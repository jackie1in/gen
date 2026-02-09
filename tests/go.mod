module github.com/jackie1in/gen/tests

go 1.16

require (
	github.com/jackie1in/gen v0.0.0-20260209125455-5ffbace9c9b6 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	gorm.io/driver/mysql v1.5.7
	gorm.io/driver/sqlite v1.4.4
	gorm.io/gen v0.3.19 // indirect
	gorm.io/gorm v1.25.12
	gorm.io/hints v1.1.1 // indirect
	gorm.io/plugin/dbresolver v1.5.3
)

replace gorm.io/gen => ../
