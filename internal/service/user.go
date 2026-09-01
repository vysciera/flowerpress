package service

import (
	"context"
	"errors"
	"strings"

	"flowerpress/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameRequired   = errors.New("username is required")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) *UserService {
	return &UserService{
		users: users,
	}
}

func (s *UserService) Register(ctx context.Context, username string, password string) (*domain.User, error) {
	username = strings.TrimSpace(username)

	if username == "" {
		return nil, ErrUsernameRequired
	}

	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	_, err := s.users.ByUsername(ctx, username)

	switch {
	case err == nil:
		return nil, ErrUsernameTaken

	case !errors.Is(err, domain.ErrUserNotFound):
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(hash),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, username string, password string) (*domain.User, error) {
	username = strings.TrimSpace(username)

	user, err := s.users.ByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
