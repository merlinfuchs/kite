package session

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

type fakeAPITokenStore struct {
	// Methods the tests don't use panic.
	store.APITokenStore

	tokens map[string]*model.APIToken
}

func (f *fakeAPITokenStore) APITokenByKeyHash(ctx context.Context, keyHash string) (*model.APIToken, error) {
	token, ok := f.tokens[keyHash]
	if !ok {
		return nil, store.ErrNotFound
	}
	copy := *token
	return &copy, nil
}

func (f *fakeAPITokenStore) UpdateAPITokenLastUsedAt(ctx context.Context, id string, lastUsedAt time.Time, interval time.Duration) error {
	for _, token := range f.tokens {
		if token.ID == id {
			token.LastUsedAt = null.TimeFrom(lastUsedAt)
		}
	}
	return nil
}

// serve sends a request through a route registered like in routes.go, in a
// group with the session middleware that allows API tokens if allowToken is
// set, and returns the status, the error code and the request's context.
func serve(t *testing.T, sessionMiddleware handler.MiddlewareFunc, allowToken bool, req *http.Request) (int, string, *handler.Context) {
	t.Helper()

	mux := http.NewServeMux()
	group := handler.Group(mux, "", sessionMiddleware)
	if allowToken {
		group = group.AllowAPIToken()
	}
	var ctx *handler.Context
	group.Get("/test", func(c *handler.Context) error {
		ctx = c
		return nil
	})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Error.Code, ctx
}

func tokenRequest(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestAPITokenSession(t *testing.T) {
	now := time.Now().UTC()
	tokenStore := &fakeAPITokenStore{tokens: map[string]*model.APIToken{
		util.HashKey("kite_valid"):   {ID: "valid", UserID: "user"},
		util.HashKey("kite_expired"): {ID: "expired", UserID: "user", ExpiresAt: null.TimeFrom(now.Add(-time.Minute))},
		util.HashKey("kite_recent"):  {ID: "recent", UserID: "user", LastUsedAt: null.TimeFrom(now.Add(-time.Second))},
	}}
	manager := NewSessionManager(SessionManagerConfig{}, &fakeSessionStore{}, tokenStore)

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantCode   string
	}{
		{"valid", "kite_valid", http.StatusOK, ""},
		{"expired", "kite_expired", http.StatusUnauthorized, "invalid_api_token"},
		{"unknown", "kite_unknown", http.StatusUnauthorized, "invalid_api_token"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, c := serve(t, manager.RequireSession, true, tokenRequest(tt.token))
			if status != tt.wantStatus || code != tt.wantCode {
				t.Fatalf("got %d %q, want %d %q", status, code, tt.wantStatus, tt.wantCode)
			}
			if status == http.StatusOK && (c.Session.UserID != "user" || c.APIToken.ID != "valid") {
				t.Errorf("got user %q and token %+v, want user and token valid", c.Session.UserID, c.APIToken)
			}
		})
	}

	if !tokenStore.tokens[util.HashKey("kite_valid")].LastUsedAt.Valid {
		t.Error("last use of the token wasn't saved")
	}
	serve(t, manager.RequireSession, true, tokenRequest("kite_recent"))
	if tokenStore.tokens[util.HashKey("kite_recent")].LastUsedAt.Time.After(now) {
		t.Error("last use was saved again within APITokenLastUsedInterval")
	}
}

func TestAPITokenNeedsAllowAPIToken(t *testing.T) {
	tokenStore := &fakeAPITokenStore{tokens: map[string]*model.APIToken{
		util.HashKey("kite_token"): {ID: "token", UserID: "user"},
	}}
	key := util.SecureKey()
	sessionStore := &fakeSessionStore{sessions: map[string]*model.Session{
		util.HashKey(key): {KeyHash: util.HashKey(key), UserID: "user", ExpiresAt: time.Now().Add(SessionExpiry)},
	}}
	manager := NewSessionManager(SessionManagerConfig{}, sessionStore, tokenStore)

	if status, code, _ := serve(t, manager.RequireSession, false, tokenRequest("kite_token")); status != http.StatusForbidden || code != "api_token_not_allowed" {
		t.Errorf("token without AllowAPIToken: got %d %q, want 403 api_token_not_allowed", status, code)
	}
	if status, _, _ := serve(t, manager.RequireSession, true, tokenRequest("kite_token")); status != http.StatusOK {
		t.Errorf("token with AllowAPIToken: got %d, want 200", status)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: key})
	if status, _, _ := serve(t, manager.RequireSession, false, req); status != http.StatusOK {
		t.Errorf("login without AllowAPIToken: got %d, want 200", status)
	}

	// Public routes work without a session, so a token there is ignored.
	status, _, c := serve(t, manager.OptionalSession, false, tokenRequest("kite_token"))
	if status != http.StatusOK || c.Session != nil || c.APIToken != nil {
		t.Errorf("token on public route: got %d with session %+v, want 200 without one", status, c.Session)
	}

	// The auth scheme is case-insensitive.
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "bearer kite_token")
	if status, _, c := serve(t, manager.RequireSession, true, req); status != http.StatusOK || c.APIToken == nil {
		t.Errorf("lowercase bearer: got %d, want 200 with the token", status)
	}
}
