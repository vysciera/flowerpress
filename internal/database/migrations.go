package database

import (
	"database/sql"
	"fmt"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

// God I love SQL.
var migrations = []Migration{
	{
		Version: 1,
		Name:    "create users",
		SQL: `
			CREATE TABLE users (
				id INTEGER PRIMARY KEY,
				username TEXT NOT NULL COLLATE NOCASE UNIQUE,
				password_hash TEXT NOT NULL,
				recovery_hash TEXT,
				session_version INTEGER NOT NULL DEFAULT 0,
				created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTMAP
			);
		`,
	},
}

func Migrate(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("initialize migrations table: %w", err)
	}

	for _, migration := range migrations {
		applied, err := migrationApplied(db, migration.Version)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", migration.Version, err)
		}

		if applied {
			continue
		}

		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("apply migration %d (%s): %w", migration.Version, migration.Name, err)
		}
	}

	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
	`)

	return err
}

func migrationApplied(db *sql.DB, version int) (bool, error) {
	var exists int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM schema_migrations
		WHERE version = ?
	`, version).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

func applyMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.Exec(migration.SQL); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO schema_migrations (version, name)
		VALUES (?, ?)
	`, migration.Version, migration.Name); err != nil {
		return err
	}

	return tx.Commit()
}
