package guild

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/api/access"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type GuildHandler struct {
	guilds        store.GuildStore
	accessManager *access.AccessManager
}

func NewHandler(guilds store.GuildStore, accessManager *access.AccessManager) *GuildHandler {
	return &GuildHandler{
		guilds:        guilds,
		accessManager: accessManager,
	}
}

func (h *GuildHandler) HandleGuildList(c *fiber.Ctx) error {
	session := session.GetSession(c)

	res := make([]wire.Guild, 0, len(session.GuildIDs))
	for _, guildID := range session.GuildIDs {
		guild, err := h.guilds.GetGuild(c.Context(), guildID)
		if err != nil {
			if err == store.ErrNotFound {
				continue
			}
			return err
		}

		perms, err := h.accessManager.GetGuildPermissionsForUser(c.Context(), guild.ID, session.UserID)
		if err != nil {
			return err
		}

		g := wire.GuildToWire(guild)
		g.UserIsOwner = perms.UserIsOwner
		g.UserPermissions = fmt.Sprintf("%d", perms.UserPermissions)
		g.BotPermissions = fmt.Sprintf("%d", perms.BotPermissions)

		res = append(res, g)
	}

	return c.JSON(wire.GuildListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *GuildHandler) HandleGuildGet(c *fiber.Ctx) error {
	// TODO: also return permissions

	guild, err := h.guilds.GetGuild(c.Context(), c.Params("guildID"))
	if err != nil {
		if err == store.ErrNotFound {
			return helpers.NotFound("unknown_guild", "Guild not found")
		}
		return err
	}

	return c.JSON(wire.GuildGetResponse{
		Success: true,
		Data:    wire.GuildToWire(guild),
	})
}
