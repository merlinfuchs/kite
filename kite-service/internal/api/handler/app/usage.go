package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

func (h *AppHandler) HandleAppUsageSummaryGet(c *fiber.Ctx) error {
	usage, err := h.appUsages.GetAppUsageSummary(c.Context(), distype.Snowflake(c.Params("appID")))
	if err != nil {
		return err
	}

	return c.JSON(wire.AppUsageSummaryGetResponse{
		Success: true,
		Data:    wire.AppUsageSummaryToWire(usage),
	})
}
