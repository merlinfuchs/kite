package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

// OAuthCode is made when a user approves an OAuth client, like Claude, and
// the client exchanges it for an API token.
type OAuthCode struct {
	CodeHash string
	UserID   string
	// The URL of the client's metadata document.
	ClientID      string
	ClientName    string
	RedirectURI   string
	CodeChallenge string
	Scopes        []string
	// The apps the token can use, nil for all apps of the user.
	AppIDs   []string
	Resource string
	// The token the code was exchanged for, null until it's used.
	APITokenID null.String
	ExpiresAt  time.Time
	CreatedAt  time.Time
}
