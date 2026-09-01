package turso

import (
	"context"
	"errors"
	"testing"
	"time"

	"flowerpress/internal/domain"
)

func testSessionRepository(t *testing.T) (*SessionRepository, *domain.User) {
	t.Helper()

	db := testDatabase(t)

	users := NewUserRepository(db)
	sessions := NewSessionRepository(db)

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

	return sessions, user
}

func TestSessionRepositoryCreate(t *testing.T) {
	repo, user := testSessionRepository(t)
	ctx := context.Background()

	session := &domain.Session{
		UserID:         user.ID,
		TokenHash:      "token-hash",
		SessionVersion: user.SessionVersion,
		ExpiresAt:      time.Now().UTC().Add(24 * time.Hour),
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	if session.ID == 0 {
		t.Fatal("expected session ID")
	}

	if session.UserID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, session.ID)
	}

	if session.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt")
	}
}

func TestSessionRepositoryByTokenHash(t *testing.T) {
	repo, user := testSessionRepository(t)
	ctx := context.Background()

	session := &domain.Session{
		UserID:         user.ID,
		TokenHash:      "abcdef123456",
		SessionVersion: user.SessionVersion,
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	found, err := repo.ByTokenHash(
		ctx,
		"abcdef123456",
	)

	if err != nil {
		t.Fatalf("find session: %v", err)
	}

	if found.ID != session.ID {
		t.Fatalf("expected ID %d, got %d", session.ID, found.ID)
	}

	if found.UserID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, found.UserID)
	}
}

func TestSessionRepositoryByTokenHashNotFound(t *testing.T) {
	repo, _ := testSessionRepository(t)

	_, err := repo.ByTokenHash(
		context.Background(),
		"does-not-exist",
	)

	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionRepositoryDelete(t *testing.T) {
	repo, user := testSessionRepository(t)
	ctx := context.Background()

	session := &domain.Session{
		UserID:         user.ID,
		TokenHash:      "delete-me",
		SessionVersion: user.SessionVersion,
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	if err := repo.Delete(ctx, session.ID); err != nil {
		t.Fatalf("delete session; %v", err)
	}

	_, err := repo.ByTokenHash(
		ctx,
		session.TokenHash,
	)

	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

// Test non-existent session
func TestSessionRepositoryDeleteNotFound(t *testing.T) {
	repo, _ := testSessionRepository(t)

	err := repo.Delete(
		context.Background(),
		999,
	)

	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionRepositoryDeleteByUserID(t *testing.T) {
	repo, user := testSessionRepository(t)
	ctx := context.Background()

	first := &domain.Session{
		UserID:         user.ID,
		TokenHash:      "first",
		SessionVersion: user.SessionVersion,
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
	}

	second := &domain.Session{
		UserID:         user.ID,
		TokenHash:      "second",
		SessionVersion: user.SessionVersion,
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first session: %v", err)
	}

	if err := repo.Create(ctx, second); err != nil {
		t.Fatalf("create second session: %v", err)
	}

	if err := repo.DeleteByUserID(ctx, user.ID); err != nil {
		t.Fatalf("delete user sessions: %v", err)
	}

	for _, tokenHash := range []string{"first", "second"} {
		_, err := repo.ByTokenHash(ctx, tokenHash)

		if !errors.Is(err, domain.ErrSessionNotFound) {
			t.Fatalf("expected session %q to be deleted, got %q", tokenHash, err)
		}
	}
}

func TestSessionRepositoryDeleteExpired(t *testing.T) {
	repo, user := testSessionRepository(t)
	ctx := context.Background()

	now := time.Now().UTC()

	expired := &domain.Session{
		UserID:         user.ID,
		TokenHash:      "expired",
		SessionVersion: user.SessionVersion,
		ExpiresAt:      now.Add(-time.Hour),
	}

	active := &domain.Session{
		UserID:         user.ID,
		TokenHash:      "active",
		SessionVersion: user.SessionVersion,
		ExpiresAt:      now.Add(time.Hour),
	}

	if err := repo.Create(ctx, expired); err != nil {
		t.Fatalf("create expired session: %v", err)
	}

	if err := repo.Create(ctx, active); err != nil {
		t.Fatalf("create active session: %v", err)
	}

	if err := repo.DeleteExpired(ctx, now); err != nil {
		t.Fatalf("delete expired sessions: %v", err)
	}

	_, err := repo.ByTokenHash(ctx, "expired")
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expect expired session to be deleted, got %v", err)
	}

	found, err := repo.ByTokenHash(ctx, "active")
	if err != nil {
		t.Fatalf("expect active session to remain: %v", err)
	}

	if found.TokenHash != "active" {
		t.Fatalf("expected active token hash, got %q", found.TokenHash)
	}
}
