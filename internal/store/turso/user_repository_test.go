package turso

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"flowerpress/internal/database"
	"flowerpress/internal/domain"
)

func testUserRepository(t *testing.T) *UserRepository {
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

	return NewUserRepository(db)
}

func TestUserReposioryCreate(t *testing.T) {
	repo := testUserRepository(t)
	ctx := context.Background()

	user := &domain.User{
		Username:     "flower",
		PasswordHash: "flowerhash",
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if user.ID == 0 {
		t.Fatal("expected user ID to be assigned")
	}

	if user.Username != "flower" {
		t.Fatalf("expected username %q, got %q", "flower", user.Username)
	}

	if user.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt to be populated")
	}

	if user.UpdatedAt.IsZero() {
		t.Fatalf("expected UpdatedAt to be populated")
	}
}

func TestUserRepositoryByUsername(t *testing.T) {
	repo := testUserRepository(t)
	ctx := context.Background()

	user := &domain.User{
		Username:     "flower",
		PasswordHash: "flowerhash",
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	found, err := repo.ByUsername(ctx, "flower")
	if err != nil {
		t.Fatalf("find user: %v", err)
	}

	if found.ID != user.ID {
		t.Fatalf("expected ID %d, got %d", user.ID, found.ID)
	}
}

func TestUserRepositoryByUsernameCaseInsensitive(t *testing.T) {
	repo := testUserRepository(t)
	ctx := context.Background()

	user := &domain.User{
		Username:     "FloWER",
		PasswordHash: "flowerhash",
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	found, err := repo.ByUsername(ctx, "FLOWER")
	if err != nil {
		t.Fatalf("find user: %v", err)
	}

	if found.ID != user.ID {
		t.Fatalf("expected ID %d, got %d", user.ID, found.ID)
	}
}

func TestUserRepositoryUpdate(t *testing.T) {
	repo := testUserRepository(t)
	ctx := context.Background()

	user := &domain.User{
		Username:     "flower",
		PasswordHash: "old-flowerhash",
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	user.Username = "garden"
	user.PasswordHash = "new-hash"
	user.SessionVersion++

	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("update user: %v", err)
	}

	found, err := repo.ByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("find updated user: %v", err)
	}

	if found.Username != "garden" {
		t.Fatalf("expected username %q, got %q", "garden", found.Username)
	}

	if found.PasswordHash != "new-hash" {
		t.Fatalf("password hash was not updated")
	}

	if found.SessionVersion != 1 {
		t.Fatalf("expected session version 1, got %d", found.SessionVersion)
	}
}

func UserRepositoryDelete(t *testing.T) {
	repo := testUserRepository(t)
	ctx := context.Background()

	user := &domain.User{
		Username:     "flower",
		PasswordHash: "flowerhash",
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := repo.Delete(ctx, user.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	_, err := repo.ByID(ctx, user.ID)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepositoryByIDNotFound(t *testing.T) {
	repo := testUserRepository(t)

	_, err := repo.ByID(context.Background(), 999)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
