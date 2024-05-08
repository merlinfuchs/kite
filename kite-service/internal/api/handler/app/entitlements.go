package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

func (h *AppHandler) HandleAppEntitlementsResolvedGet(c *fiber.Ctx) error {
	entitlements, err := h.AppEntitlements.GetResolvedAppEntitlement(c.Context(), distype.Snowflake(c.Params("appID")))
	if err != nil {
		return err
	}

	return c.JSON(wire.AppEntitlementResolvedGetResponse{
		Success: true,
		Data:    wire.AppEntitlementResolvedToWire(entitlements),
	})
}
