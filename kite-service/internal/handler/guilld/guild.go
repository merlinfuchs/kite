package guild

import (
	"github.com/gofiber/fiber/v2"
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
