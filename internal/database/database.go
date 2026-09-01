package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "turso.tech/database/tursogo"
)

func Open(path string) (*sql.DB, error) {
	if err := ensureDirectory(path); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("turso", path)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return db, nil
}

func ensureDirectory(path string) error {
	dir := filepath.Dir(path)

	if dir == "." {
		return nil
	}

	return os.MkdirAll(dir, 0755)
}
