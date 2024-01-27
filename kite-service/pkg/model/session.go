package model

import "time"

type Session struct {
	TokenHash   string
	UserID      string
	GuildIds    []string
	AccessToken string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}
