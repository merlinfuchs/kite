package session

import (
	"context"
	"errors"
	"fmt"
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

	sameSite := http.SameSiteNoneMode
	if s.config.StrictCookies {
		sameSite = http.SameSiteStrictMode
	}

	c.SetCookie(&http.Cookie{
		Name:     SessionCookieName,
		Value:    key,
		Secure:   s.config.SecureCookies,
		HttpOnly: true,
		SameSite: sameSite,
		MaxAge:   int(SessionExpiry.Seconds()),
		Path:     "/",
	})

	return key, session, nil
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

	return session, nil
}
