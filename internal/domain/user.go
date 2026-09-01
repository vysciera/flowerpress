package domain

import "time"

type User struct {
	ID             int64
	Username       string
	PasswordHash   string
	RecoveryHash   *string // DB Column is nullable
	SessionVersion int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
