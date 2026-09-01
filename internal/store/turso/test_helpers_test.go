package turso

import (
	"database/sql"
	"path/filepath"
	"testing"

	"flowerpress/internal/database"
)

func testDatabase(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(
		t.TempDir(),
		"flowerpress.db",
	)

	db, err := database.Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	return db
}
