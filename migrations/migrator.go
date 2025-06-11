package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed */*.sql
var embedMigrations embed.FS

// Performs migrations
func Migrate(db *sql.DB, driverName string) error {
	const op = "migrations.Migrate"

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("%s: failed to set dialect: %w", op, err)
	}

	if err := goose.Up(db, driverName); err != nil {
		return fmt.Errorf("%s: failed to apply migrations: %w", op, err)
	}
	return nil
}
