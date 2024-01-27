package access

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
)

func (m *AccessManager) CheckGuildAccessForRequest(c *fiber.Ctx, guildID string) error {
	session := session.GetSession(c)

	perms, err := m.GetGuildPermissionsForUser(c.Context(), session.UserID, guildID)
	if err != nil {
		return err
	}

	if !perms.UserIsMember {
		return helpers.Forbidden("missing_access", "You are not a member of this guild")
	}

	if !perms.BotIsMember {
		return helpers.Forbidden("bot_missing_access", "The bot is not a member of this guild")
	}

	if !perms.UserIsOwner && perms.UserPermissions&8 == 0 {
		return helpers.Forbidden("missing_access", "You don't have administrator permissions in this guild")
	}

	return nil
}
