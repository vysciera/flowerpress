package database

import (
	"path/filepath"
	"testing"
)

func TestOpenEnablesForeignKeys(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"flowerpress.db",
	)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	defer db.Close()
	var enabled int

	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&enabled); err != nil {
		t.Fatalf("query foreign_keys pragma: %v", err)
	}

	if enabled != 1 {
		t.Fatalf("expected foreign keys to be enabled, got %d", enabled)
	}
}

func TestOpenEnforcesForeignKeys(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"flowerpress.db",
	)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	defer db.Close()

	// Slight overkill
	_, err = db.Exec(`
		CREATE TABLE parents (
			id INTEGER PRIMARY KEY
		);

		CREATE TABLE children (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER NOT NULL,

			FOREIGN KEY (parent_id)
				REFERENCES parents(id)
		)
	`)

	if err != nil {
		t.Fatalf("create tables: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO children (parent_id)
		VALUES (999)
	`)

	if err == nil {
		t.Fatalf("expect foreign key constrain violation")
	}
}
