package session

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
)

func (m *SessionManager) RequireSession(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		session, err := m.Session(c)
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}

		if session == nil {
			return handler.ErrUnauthorized("unauthorized", "Session required")
		}

		c.Session = session
		return next(c)
	}
}

func (m *SessionManager) OptionalSession(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		session, err := m.Session(c)
		if err != nil {
			return err
		}

		c.Session = session
		return next(c)
	}
}

// RequireLogin is for what an API token can't do, like managing API tokens or
// deleting an app. It comes after RequireSession.
func (m *SessionManager) RequireLogin(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		if c.Session == nil || c.Session.APITokenID != "" {
			return handler.ErrForbidden("login_required", "This needs a login, an API token can't do it")
		}
		return next(c)
	}
}
