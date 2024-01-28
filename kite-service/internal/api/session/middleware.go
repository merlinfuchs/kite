package session

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
)

type SessionMiddleware struct {
	manager *SessionManager
}

func NewMiddleware(manager *SessionManager) *SessionMiddleware {
	return &SessionMiddleware{
		manager: manager,
	}
}

func (m *SessionMiddleware) SessionRequired() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		session, err := m.manager.GetSession(c)
		if err != nil {
			return err
		}

		if session == nil {
			return helpers.Unauthorized("invalid_session", "No valid session, try logging in again.")
		}

		if session.ExpiresAt.Before(time.Now().UTC()) {
			return helpers.Unauthorized("invalid_session", "Session expired, try logging in again.")
		}

		c.Locals("session", session)

		return c.Next()
	}
}

func (m *SessionMiddleware) SessionOptional() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		session, err := m.manager.GetSession(c)
		if err != nil {
			slog.With(logattr.Error(err)).Error("Failed to validate session")
		}

		c.Locals("session", session)

		return c.Next()
	}
}

func GetSession(c *fiber.Ctx) *Session {
	return c.Locals("session").(*Session)
}
