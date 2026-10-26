package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

// APITokenScopeAll lets a token do everything its user can through the API,
// except what needs a login. It's the only scope for now.
const APITokenScopeAll = "all"

type APIToken struct {
	ID         string
	UserID     string
	Name       string
	KeyHash    string
	Scopes     []string
	LastUsedAt null.Time
	ExpiresAt  null.Time
	CreatedAt  time.Time
}
