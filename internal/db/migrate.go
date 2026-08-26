package db

import (
	"database/sql"
	"embed"
	"fmt"
)

//go:embed schema/000001_init_schema.up.sql
var migrationFS embed.FS

func Migrate(database *sql.DB) error {
	migration, err := migrationFS.ReadFile("schema/000001_init_schema.up.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	if _, err := database.Exec(string(migration)); err != nil {
		return fmt.Errorf("run migration: %w", err)
	}
	return nil
}
