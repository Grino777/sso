package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed */*.sql
var embedMigrations embed.FS

// Performs migrations
func Migrate(db *sql.DB, driverName string) {

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic("failed to set dialect: " + err.Error())
	}

	if err := goose.Up(db, driverName); err != nil {
		panic("failed to apply migrations: " + err.Error())
	}
}
