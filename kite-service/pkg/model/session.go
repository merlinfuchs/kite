package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type SessionType string

const (
	SessionTypeWebApp SessionType = "WEB_APP"
	SessionTypeCLI    SessionType = "CLI"
)

type Session struct {
	TokenHash   string
	Type        SessionType
	UserID      string
	GuildIds    []string
	AccessToken string
	Revoked     bool
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type PendingSession struct {
	Code      string
	Token     null.String
	CreatedAt time.Time
	ExpiresAt time.Time
}
