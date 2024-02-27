package guild

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

func (h *GuildHandler) HandleGuildUsageSummaryGet(c *fiber.Ctx) error {
	usage, err := h.guildUsages.GetGuildUsageSummary(c.Context(), distype.Snowflake(c.Params("guildID")))
	if err != nil {
		return err
	}

	return c.JSON(wire.GuildUsageSummaryGetResponse{
		Success: true,
		Data:    wire.GuildUsageSummaryToWire(usage),
	})
}
