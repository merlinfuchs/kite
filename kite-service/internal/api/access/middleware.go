package access

import (
	"github.com/gofiber/fiber/v2"
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

func (m *AccessMiddleware) GuildAccessRequired() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		session := session.GetSession(c)
		guildID := c.Params("guildID")

		perms, err := m.manager.GetGuildPermissionsForUser(c.Context(), guildID, session.UserID)
		if err != nil {
			return err
		}

		if !perms.BotIsMember {
			return helpers.Forbidden("bot_missing_access", "The bot is not a member of this guild")
		}

		if !perms.UserIsMember {
			return helpers.Forbidden("missing_access", "You are not a member of this guild")
		}

		if !perms.UserIsOwner && perms.UserPermissions&8 == 0 {
			return helpers.Forbidden("missing_access", "You don't have administrator permissions in this guild")
		}

		c.Locals("guildPermissions", perms)
		return c.Next()
	}
}
