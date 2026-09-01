package httpapi

import (
	"path/filepath"
	"testing"
	"time"

	"flowerpress/internal/database"
	"flowerpress/internal/service"
	"flowerpress/internal/store/turso"
)

func testServer(t *testing.T) *Server {
	t.Helper()

	path := filepath.Join(
		t.TempDir(),
		"flowerpress.db",
	)

	db, err := database.Open(path)
	if err != nil {
		t.Fatalf("open databse: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	users := turso.NewUserRepository(db)
	sessions := turso.NewSessionRepository(db)

	return NewServer(
		service.NewUserService(users),
		service.NewSessionService(sessions, users, 24*time.Hour),
	)
}
