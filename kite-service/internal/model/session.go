package model

import "time"

type Session struct {
	KeyHash   string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	// Set when the request was made with an API token instead of a login.
	// Such sessions aren't stored.
	APITokenID string
}
