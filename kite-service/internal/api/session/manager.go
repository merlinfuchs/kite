package session

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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
	// APITokenLastUsedInterval is how often the last use of a token is saved,
	// so requests in a row don't each write it.
	APITokenLastUsedInterval = time.Minute
)

type SessionManagerConfig struct {
	StrictCookies bool
	SecureCookies bool
}

type SessionManager struct {
	config        SessionManagerConfig
	sessionStore  store.SessionStore
	apiTokenStore store.APITokenStore
}

func NewSessionManager(
	config SessionManagerConfig,
	sessionStore store.SessionStore,
	apiTokenStore store.APITokenStore,
) *SessionManager {
	return &SessionManager{
		config:        config,
		sessionStore:  sessionStore,
		apiTokenStore: apiTokenStore,
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

// Session is the session of the request's login, or of its API token if it
// has one in the Authorization header. Routes that don't allow API tokens
// reject them before anything else runs.
func (s *SessionManager) Session(c *handler.Context) (*model.Session, error) {
	if token, ok := bearerToken(c); ok {
		if !c.APITokenAllowed() {
			return nil, handler.ErrForbidden("api_token_not_allowed", "API tokens can't use this endpoint, it needs a login")
		}
		return s.apiTokenSession(c, token)
	}

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

// bearerToken is the API token in the Authorization header, if it has one.
func bearerToken(c *handler.Context) (string, bool) {
	scheme, token, ok := strings.Cut(c.Header("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	return strings.TrimSpace(token), true
}

func (s *SessionManager) apiTokenSession(c *handler.Context, token string) (*model.Session, error) {
	// A client that sends a token wants to use it, so a bad one is an error
	// instead of a request without a session.
	errInvalid := handler.ErrUnauthorized("invalid_api_token", "The API token is invalid, expired or was deleted")

	apiToken, err := s.apiTokenStore.APITokenByKeyHash(c.Context(), util.HashKey(token))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, errInvalid
		}
		return nil, fmt.Errorf("failed to get API token: %w", err)
	}

	now := time.Now().UTC()
	if apiToken.ExpiresAt.Valid && apiToken.ExpiresAt.Time.Before(now) {
		return nil, errInvalid
	}

	if !apiToken.LastUsedAt.Valid || now.Sub(apiToken.LastUsedAt.Time) > APITokenLastUsedInterval {
		if err := s.apiTokenStore.UpdateAPITokenLastUsedAt(c.Context(), apiToken.ID, now, APITokenLastUsedInterval); err != nil {
			slog.Error(
				"Failed to update API token last use",
				slog.String("api_token_id", apiToken.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	// The session isn't stored, it only says whose token it is.
	c.APIToken = apiToken
	return &model.Session{UserID: apiToken.UserID}, nil
}
