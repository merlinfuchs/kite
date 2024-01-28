package model

import "time"

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
