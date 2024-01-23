package guild

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type GuildHandler struct {
	engine *engine.PluginEngine
	guilds store.GuildStore
}

func NewHandler(engine *engine.PluginEngine, guilds store.GuildStore) *GuildHandler {
	return &GuildHandler{
		engine: engine,
		guilds: guilds,
	}
}

func (h *GuildHandler) HandleGuildList(c *fiber.Ctx) error {
	guilds, err := h.guilds.GetGuilds(c.Context())
	if err != nil {
		return err
	}

	res := make([]wire.Guild, len(guilds))
	for i, guild := range guilds {
		res[i] = wire.GuildToWire(&guild)
	}

	return c.JSON(wire.GuildListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *GuildHandler) HandleGuildGet(c *fiber.Ctx) error {
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
