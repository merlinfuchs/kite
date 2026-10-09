package session

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

const (
	SessionCookieName = "kite-session"
	SessionExpiry     = 7 * 24 * time.Hour
	// SessionRefreshInterval is how often the expiry of a session in use is pushed back.
	SessionRefreshInterval = 24 * time.Hour
)

type SessionManagerConfig struct {
	StrictCookies bool
	SecureCookies bool
}

type SessionManager struct {
	config       SessionManagerConfig
	sessionStore store.SessionStore
}

func NewSessionManager(config SessionManagerConfig, sessionStore store.SessionStore) *SessionManager {
	return &SessionManager{
		config:       config,
		sessionStore: sessionStore,
	}
}

func (s *SessionManager) CreateSessionCookie(c *handler.Context, userID string) (string, *model.Session, error) {
	key, session, err := s.CreateSession(c.Context(), userID)
	if err != nil {
		return "", nil, err
	}

	s.setSessionCookie(c, key)
	return key, session, nil
}

func (s *SessionManager) setSessionCookie(c *handler.Context, key string) {
	sameSite := http.SameSiteNoneMode
	if s.config.StrictCookies {
		sameSite = http.SameSiteStrictMode
	}

	// Keep shared caches from storing the session cookie
	c.SetHeader("Cache-Control", "private")
	c.SetCookie(&http.Cookie{
		Name:     SessionCookieName,
		Value:    key,
		Secure:   s.config.SecureCookies,
		HttpOnly: true,
		SameSite: sameSite,
		MaxAge:   int(SessionExpiry.Seconds()),
		Path:     "/",
	})
}

func (s *SessionManager) CreateSession(ctx context.Context, userID string) (string, *model.Session, error) {
	key := util.SecureKey()
	keyHash := util.HashKey(key)

	session, err := s.sessionStore.CreateSession(ctx, &model.Session{
		KeyHash:   keyHash,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(SessionExpiry),
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to create session: %w", err)
	}

	return key, session, nil
}

func (s *SessionManager) DeleteSession(c *handler.Context) error {
	defer c.DeleteCookie(SessionCookieName)

	key := c.Cookie(SessionCookieName)
	if key == "" {
		return nil
	}

	keyHash := util.HashKey(key)
	if err := s.sessionStore.DeleteSession(c.Context(), keyHash); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (s *SessionManager) Session(c *handler.Context) (*model.Session, error) {
	key := c.Cookie(SessionCookieName)
	if key == "" {
		return nil, nil
	}

	keyHash := util.HashKey(key)

	session, err := s.sessionStore.Session(c.Context(), keyHash)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	now := time.Now().UTC()
	if session.ExpiresAt.Before(now) {
		return nil, nil
	}

	expiresAt := now.Add(SessionExpiry)
	if expiresAt.Sub(session.ExpiresAt) < SessionRefreshInterval {
		return session, nil
	}

	if err := s.sessionStore.UpdateSessionExpiry(c.Context(), keyHash, expiresAt); err != nil {
		// Deleted by a concurrent logout
		if errors.Is(err, store.ErrNotFound) {
			return nil, nil
		}
		slog.Error(
			"Failed to refresh session expiry",
			slog.String("user_id", session.UserID),
			slog.String("error", err.Error()),
		)
		return session, nil
	}

	session.ExpiresAt = expiresAt
	s.setSessionCookie(c, key)
	return session, nil
}
