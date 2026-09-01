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
		t.Fatalf("expect foreign keys to be enabled, got %d", enabled)
	}
}
