package guild

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

func (h *GuildHandler) HandleGuildEntitlementsResolvedGet(c *fiber.Ctx) error {
	entitlements, err := h.guildEntitlements.GetResolvedGuildEntitlement(c.Context(), distype.Snowflake(c.Params("guildID")))
	if err != nil {
		return err
	}

	return c.JSON(wire.GuildEntitlementResolvedGetResponse{
		Success: true,
		Data:    wire.GuildEntitlementResolvedToWire(entitlements),
	})
}
