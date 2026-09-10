package domain

import (
	"context"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error

	ByID(ctx context.Context, id int64) (*User, error)
	ByUsername(ctx context.Context, username string) (*User, error)

	HasAny(ctx context.Context) (bool, error)
}
