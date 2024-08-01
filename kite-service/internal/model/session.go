package model

import "time"

type Session struct {
	KeyHash   string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}
