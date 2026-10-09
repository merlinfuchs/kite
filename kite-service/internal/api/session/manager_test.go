package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type fakeSessionStore struct {
	sessions map[string]*model.Session
}

func (f *fakeSessionStore) CreateSession(ctx context.Context, session *model.Session) (*model.Session, error) {
	f.sessions[session.KeyHash] = session
	return session, nil
}

func (f *fakeSessionStore) DeleteSession(ctx context.Context, keyHash string) error {
	delete(f.sessions, keyHash)
	return nil
}

func (f *fakeSessionStore) Session(ctx context.Context, keyHash string) (*model.Session, error) {
	session, ok := f.sessions[keyHash]
	if !ok {
		return nil, store.ErrNotFound
	}
	copy := *session
	return &copy, nil
}

func (f *fakeSessionStore) UpdateSessionExpiry(ctx context.Context, keyHash string, expiresAt time.Time) error {
	f.sessions[keyHash].ExpiresAt = expiresAt
	return nil
}

func TestSessionExpiry(t *testing.T) {
	tests := []struct {
		name        string
		expiresIn   time.Duration
		wantStatus  int
		wantRefresh bool
	}{
		{"expired", -time.Minute, http.StatusUnauthorized, false},
		{"recently refreshed", SessionExpiry - time.Hour, http.StatusOK, false},
		{"due for refresh", SessionExpiry - SessionRefreshInterval - time.Hour, http.StatusOK, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := util.SecureKey()
			keyHash := util.HashKey(key)
			expiresAt := time.Now().UTC().Add(tt.expiresIn)

			sessionStore := &fakeSessionStore{sessions: map[string]*model.Session{
				keyHash: {KeyHash: keyHash, UserID: "user", ExpiresAt: expiresAt},
			}}
			manager := NewSessionManager(SessionManagerConfig{}, sessionStore)

			h := handler.APIHandler(manager.RequireSession(func(c *handler.Context) error {
				c.SetHeader("Content-Type", "text/plain")
				return nil
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: key})
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			refreshed := sessionStore.sessions[keyHash].ExpiresAt.After(expiresAt)
			if refreshed != tt.wantRefresh {
				t.Errorf("refreshed = %v, want %v", refreshed, tt.wantRefresh)
			}

			setCookie := rec.Header().Get("Set-Cookie") != ""
			if setCookie != tt.wantRefresh {
				t.Errorf("set cookie = %v, want %v", setCookie, tt.wantRefresh)
			}
		})
	}
}
