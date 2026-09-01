package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"flowerpress/internal/domain"
	"flowerpress/internal/store/turso"
)

func testSessionService(t *testing.T, duration time.Duration) (*SessionService, *turso.UserRepository, *turso.SessionRepository, *domain.User) {
	t.Helper()

	db := testDatabase(t)

	users := turso.NewUserRepository(db)
	sessions := turso.NewSessionRepository(db)

	user := &domain.User{
		Username:     "flower",
		PasswordHash: "flowerhash",
	}

	if err := users.Create(
		context.Background(),
		user,
	); err != nil {
		t.Fatalf("create test user: %v", err)
	}

	service := NewSessionService(
		sessions,
		users,
		duration,
	)

	return service, users, sessions, user
}

func TestSessionServiceCreate(t *testing.T) {
	service, _, sessions, user := testSessionService(t, 24*time.Hour)
	ctx := context.Background()

	token, err := service.Create(ctx, user)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if token == "" {
		t.Fatalf("expected session token")
	}

	session, err := sessions.ByTokenHash(
		ctx,
		hashSessionToken(token),
	)

	if err != nil {
		t.Fatalf("find created session: %v", err)
	}

	if session.UserID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, session.UserID)
	}

	if session.SessionVersion != user.SessionVersion {
		t.Fatalf("expected session version %d, got %d", user.SessionVersion, session.SessionVersion)
	}
}

func TestSessionServiceCreateStoresTokenHash(t *testing.T) {
	service, _, sessions, user := testSessionService(t, 24*time.Hour)
	ctx := context.Background()

	token, err := service.Create(ctx, user)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	_, err = sessions.ByTokenHash(ctx, token)

	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected raw token not to be stored, got %v", err)
	}

	session, err := sessions.ByTokenHash(
		ctx,
		hashSessionToken(token),
	)

	if err != nil {
		t.Fatalf("find hashed token: %v", err)
	}

	if session.TokenHash == token {
		t.Fatal("session stored raw token")
	}
}

func TestSessionServiceValidate(t *testing.T) {
	service, _, _, user := testSessionService(t, 24*time.Hour)
	ctx := context.Background()

	token, err := service.Create(ctx, user)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	found, err := service.Validate(ctx, token)
	if err != nil {
		t.Fatalf("validate session: %v", err)
	}

	if found.ID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, found.ID)
	}
}

func TestSessionServiceValidateUnknownToken(t *testing.T) {
	service, _, _, _ := testSessionService(t, 24*time.Hour)

	_, err := service.Validate(
		context.Background(),
		"this-token-does-not-exist-xd",
	)

	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrinvalidSession, got %v", err)
	}
}

func TestSessionServiceValidateEmptyToken(t *testing.T) {
	service, _, _, _ := testSessionService(t, 24*time.Hour)

	_, err := service.Validate(
		context.Background(),
		"",
	)

	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}
}

func TestSessionServiceValidateExpiredSession(t *testing.T) {
	service, _, sessions, user := testSessionService(t, -time.Hour)
	ctx := context.Background()

	token, err := service.Create(ctx, user)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	_, err = service.Validate(ctx, token)

	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}

	_, err = sessions.ByTokenHash(
		ctx,
		hashSessionToken(token),
	)

	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected expired session to be deleted, got %v", err)
	}
}

func TestSessionServiceValidateRejectsStaleSessionVersion(t *testing.T) {
	service, users, sessions, user := testSessionService(t, 24*time.Hour)
	ctx := context.Background()

	token, err := service.Create(ctx, user)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	user.SessionVersion++

	if err := users.Update(ctx, user); err != nil {
		t.Fatalf("update user: %v", err)
	}

	_, err = service.Validate(ctx, token)

	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}

	_, err = sessions.ByTokenHash(
		ctx,
		hashSessionToken(token),
	)

	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected stale session to be deleted, got %v", err)
	}
}

func TestSessionServiceDelete(t *testing.T) {
	service, _, sessions, user := testSessionService(t, 24*time.Hour)
	ctx := context.Background()

	token, err := service.Create(ctx, user)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if err := service.Delete(ctx, token); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	_, err = sessions.ByTokenHash(
		ctx,
		hashSessionToken(token),
	)

	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("exected session to be deleted, got %v", err)
	}
}

func TestSessionServiceDeleteIsIdempotent(t *testing.T) {
	service, _, _, user := testSessionService(t, 24*time.Hour)
	ctx := context.Background()

	token, err := service.Create(ctx, user)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if err := service.Delete(ctx, token); err != nil {
		t.Fatalf("first delete: %v", err)
	}

	if err := service.Delete(ctx, token); err != nil {
		t.Fatalf("second delete: %v", err)
	}
}

func TestSessionServiceDeleteEmptyToken(t *testing.T) {
	service, _, _, _ := testSessionService(t, 24*time.Hour)

	if err := service.Delete(
		context.Background(),
		"",
	); err != nil {
		t.Fatalf("delete empty token: %v", err)
	}
}
