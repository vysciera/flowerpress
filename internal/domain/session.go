package domain

import "time"

type Session struct {
	ID             int64
	UserID         int64
	TokenHash      string
	SessionVersion int64
	ExpiresAt      time.Time
	CreatedAt      time.Time
}
