package httpapi

import (
	"context"

	"flowerpress/internal/domain"
)

type contextKey string

const userContextKey contextKey = "user"

func withUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(
		ctx,
		userContextKey,
		user,
	)
}

func userFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(userContextKey).(*domain.User)

	return user, ok
}
