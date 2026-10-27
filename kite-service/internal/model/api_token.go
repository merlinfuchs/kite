package model

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"gopkg.in/guregu/null.v4"
)

const (
	APITokenResourceKite = "kite"

	APITokenAccessRead  = "read"
	APITokenAccessWrite = "write"
)

// Scopes are "<resource>:<access>". Write access includes read.
const (
	APITokenScopeKiteRead  = APITokenResourceKite + ":" + APITokenAccessRead
	APITokenScopeKiteWrite = APITokenResourceKite + ":" + APITokenAccessWrite
)

type APIToken struct {
	ID      string
	UserID  string
	Name    string
	KeyHash string
	Scopes  []string
	// The apps the token can use, nil for all apps of its user.
	AppIDs     []string
	LastUsedAt null.Time
	ExpiresAt  null.Time
	CreatedAt  time.Time
}

// HasAccess reports whether the token's scopes give it access to resource.
func (t *APIToken) HasAccess(resource string, access string) bool {
	for _, scope := range t.Scopes {
		r, a, _ := strings.Cut(scope, ":")
		if r == resource && (a == access || a == APITokenAccessWrite) {
			return true
		}
	}
	return false
}

func (t *APIToken) CanUseApp(appID string) bool {
	return t.AppIDs == nil || slices.Contains(t.AppIDs, appID)
}

// APITokenAccessForMethod is the access a request with the HTTP method needs.
func APITokenAccessForMethod(method string) string {
	if method == http.MethodGet || method == http.MethodHead {
		return APITokenAccessRead
	}
	return APITokenAccessWrite
}
