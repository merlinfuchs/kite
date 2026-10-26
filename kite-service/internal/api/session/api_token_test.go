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

func (f *fakeAPITokenStore) UpdateAPITokenLastUsedAt(ctx context.Context, id string, lastUsedAt time.Time) error {
	for _, token := range f.tokens {
		if token.ID == id {
			token.LastUsedAt = null.TimeFrom(lastUsedAt)
		}
	}
	return nil
}

// serve runs a request with the token through RequireSession and the given
// middleware, and returns the status, the error code and the session.
func serve(t *testing.T, manager *SessionManager, token string, middleware ...handler.MiddlewareFunc) (int, string, *model.Session) {
	t.Helper()

	var session *model.Session
	f := func(c *handler.Context) error {
		session = c.Session
		return nil
	}
	for i := len(middleware) - 1; i >= 0; i-- {
		f = middleware[i](f)
	}
	h := handler.APIHandler(manager.RequireSession(f))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Error.Code, session
}

func TestAPITokenSession(t *testing.T) {
	now := time.Now().UTC()
	valid := "kite_valid"
	expired := "kite_expired"
	recent := "kite_recent"

	tokenStore := &fakeAPITokenStore{tokens: map[string]*model.APIToken{
		util.HashKey(valid):   {ID: "valid", UserID: "user"},
		util.HashKey(expired): {ID: "expired", UserID: "user", ExpiresAt: null.TimeFrom(now.Add(-time.Minute))},
		util.HashKey(recent):  {ID: "recent", UserID: "user", LastUsedAt: null.TimeFrom(now.Add(-time.Second))},
	}}
	manager := NewSessionManager(SessionManagerConfig{}, &fakeSessionStore{}, tokenStore)

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantCode   string
	}{
		{"valid", valid, http.StatusOK, ""},
		{"expired", expired, http.StatusUnauthorized, "invalid_api_token"},
		{"unknown", "kite_unknown", http.StatusUnauthorized, "invalid_api_token"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, session := serve(t, manager, tt.token)
			if status != tt.wantStatus || code != tt.wantCode {
				t.Fatalf("got %d %q, want %d %q", status, code, tt.wantStatus, tt.wantCode)
			}
			if status == http.StatusOK && (session.UserID != "user" || session.APITokenID != "valid") {
				t.Errorf("session = %+v, want user and token valid", session)
			}
		})
	}

	if !tokenStore.tokens[util.HashKey(valid)].LastUsedAt.Valid {
		t.Error("last use of the token wasn't saved")
	}
	serve(t, manager, recent)
	if tokenStore.tokens[util.HashKey(recent)].LastUsedAt.Time.After(now) {
		t.Error("last use was saved again within APITokenLastUsedInterval")
	}
}

func TestRequireLogin(t *testing.T) {
	token := "kite_token"
	tokenStore := &fakeAPITokenStore{tokens: map[string]*model.APIToken{
		util.HashKey(token): {ID: "token", UserID: "user"},
	}}
	key := util.SecureKey()
	sessionStore := &fakeSessionStore{sessions: map[string]*model.Session{
		util.HashKey(key): {KeyHash: util.HashKey(key), UserID: "user", ExpiresAt: time.Now().Add(SessionExpiry)},
	}}
	manager := NewSessionManager(SessionManagerConfig{}, sessionStore, tokenStore)

	if status, code, _ := serve(t, manager, token, manager.RequireLogin); status != http.StatusForbidden || code != "login_required" {
		t.Errorf("token: got %d %q, want 403 login_required", status, code)
	}

	h := handler.APIHandler(manager.RequireSession(manager.RequireLogin(func(c *handler.Context) error {
		return nil
	})))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: key})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("login: got %d, want 200", rec.Code)
	}
}
