package session

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
)

type SessionMiddleware struct {
	manager *SessionManager
}

func NewSessionMiddleware(manager *SessionManager) *SessionMiddleware {
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
			return helpers.Unauthorized("invalid_session", "No valid session, perhaps it expired, try logging in again.")
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
