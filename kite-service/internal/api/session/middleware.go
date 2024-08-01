package session

import (
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
)

func (m *SessionManager) RequireSession(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		session, err := m.Session(c)
		if err != nil {
			return err
		}

		if session == nil {
			return handler.ErrUnauthorized("unauthorized", "Session required")
		}

		c.Session = session
		return next(c)
	}
}
