package access

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
)

type AccessMiddleware struct {
	manager *AccessManager
}

func NewMiddleware(manager *AccessManager) *AccessMiddleware {
	return &AccessMiddleware{
		manager: manager,
	}
}

func (m *AccessMiddleware) AppAccessRequired() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		session := session.GetSession(c)
		appID := c.Params("appID")

		perms, err := m.manager.GetAppPermissionsForUser(c.Context(), distype.Snowflake(appID), session.UserID)
		if err != nil {
			return err
		}

		if !perms.UserIsOwner {
			return helpers.Forbidden("missing_access", "You don't have access to this app")
		}

		c.Locals("appPermissions", perms)
		return c.Next()
	}
}
