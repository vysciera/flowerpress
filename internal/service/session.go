package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"flowerpress/internal/domain"
)

var ErrInvalidSession = errors.New("invalid session")

type SessionService struct {
	sessions domain.SessionRepository
	users    domain.UserRepository
	duration time.Duration
}

func NewSessionService(sessions domain.SessionRepository, users domain.UserRepository, duration time.Duration) *SessionService {
	return &SessionService{
		sessions: sessions,
		users:    users,
		duration: duration,
	}
}

func generateSessionToken() (string, error) {
	token := make([]byte, 32)

	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(token), nil
}

func hashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}

func (s *SessionService) Create(ctx context.Context, user *domain.User) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	// Raw token never enters domain.Session
	session := &domain.Session{
		UserID:         user.ID,
		TokenHash:      hashSessionToken(token),
		SessionVersion: user.SessionVersion,
		ExpiresAt:      time.Now().UTC().Add(s.duration),
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return "", err
	}

	return token, nil
}

func (s *SessionService) Validate(ctx context.Context, token string) (*domain.User, error) {
	if token == "" {
		return nil, ErrInvalidSession
	}

	tokenHash := hashSessionToken(token)

	session, err := s.sessions.ByTokenHash(
		ctx,
		tokenHash,
	)

	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return nil, ErrInvalidSession
		}

		return nil, err
	}

	if !session.ExpiresAt.After(time.Now().UTC()) {
		_ = s.sessions.Delete(ctx, session.ID)

		return nil, ErrInvalidSession
	}

	user, err := s.users.ByID(
		ctx,
		session.UserID,
	)

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, ErrInvalidSession
		}

		return nil, err
	}

	if session.SessionVersion != user.SessionVersion {
		_ = s.sessions.Delete(ctx, session.ID)

		return nil, ErrInvalidSession
	}

	return user, nil
}

func (s *SessionService) Delete(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	session, err := s.sessions.ByTokenHash(
		ctx,
		hashSessionToken(token),
	)

	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return nil
		}

		return err
	}

	return s.sessions.Delete(ctx, session.ID)
}

func (s *SessionService) DeleteExpired(ctx context.Context) error {
	return s.sessions.DeleteExpired(
		ctx,
		time.Now().UTC(),
	)
}
