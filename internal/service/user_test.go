package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"flowerpress/internal/database"
	"flowerpress/internal/store/turso"
)

func testUserService(t *testing.T) *UserService {
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
		t.Fatalf("migrate databse: %v", err)
	}

	repo := turso.NewUserRepository(db)
	return NewUserService(repo)
}

func TestUserServiceRegister(t *testing.T) {
	users := testUserService(t)

	user, err := users.Register(
		context.Background(),
		"flower",
		"newgarden",
	)

	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	if user.ID == 0 {
		t.Fatal("expected user ID")
	}

	if user.Username != "flower" {
		t.Fatalf("expected username %q, got %q", "flower", user.Username)
	}

	if user.PasswordHash == "newgarden" {
		t.Fatal("password was stored in plaintext")
	}
}

func TestUserServiceRegisterDuplicateUsername(t *testing.T) {
	users := testUserService(t)
	ctx := context.Background()

	_, err := users.Register(
		ctx,
		"flower",
		"newgarden",
	)

	if err != nil {
		t.Fatalf("register first user: %v", err)
	}

	_, err = users.Register(
		ctx,
		"FLOWER",
		"newergarden",
	)

	if !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestUserServiceAuthenticate(t *testing.T) {
	users := testUserService(t)
	ctx := context.Background()

	created, err := users.Register(
		ctx,
		"flower",
		"newgarden",
	)

	if err != nil {
		t.Fatalf("register user: %v", err) // Redundant (?)
	}

	authenticated, err := users.Authenticate(
		ctx,
		"flower",
		"newgarden",
	)

	if err != nil {
		t.Fatalf("authenticate user: %v", err)
	}

	if authenticated.ID != created.ID {
		t.Fatalf("expected user ID %d, got %d", created.ID, authenticated.ID)
	}
}

func TestUserServiceAuthenticateWrongPassword(t *testing.T) {
	users := testUserService(t)
	ctx := context.Background()

	_, err := users.Register(
		ctx,
		"flower",
		"newgarden",
	)

	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	_, err = users.Authenticate(
		ctx,
		"flower",
		"badpwd-garden",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserServiceAuthenticateUnknownUser(t *testing.T) {
	users := testUserService(t)

	_, err := users.Authenticate(
		context.Background(),
		"nobody",
		"newgarden",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserServiceRegisterRequiresUsername(t *testing.T) {
	users := testUserService(t)

	_, err := users.Register(
		context.Background(),
		"   ",
		"newgarden",
	)

	if !errors.Is(err, ErrUsernameRequired) {
		t.Fatalf("expected ErrUsernameRequired, got %v", err)
	}
}

func TestUserServiceRegisterRequiresPasswordLength(t *testing.T) {
	users := testUserService(t)

	_, err := users.Register(
		context.Background(),
		"flower",
		"tiny",
	)

	if !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}
