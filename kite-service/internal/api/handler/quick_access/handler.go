package quickaccess

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type QuickAccessHandler struct {
	items store.QuickAccessStore
}

func NewHandler(items store.QuickAccessStore) *QuickAccessHandler {
	return &QuickAccessHandler{
		items: items,
	}
}

func (h *QuickAccessHandler) HandleQuickAccessItemList(c *fiber.Ctx) error {
	items, err := h.items.GetQuickAccessItems(c.Context(), c.Params("appID"), 3)
	if err != nil {
		return err
	}

	res := make([]wire.QuickAccessItem, len(items))
	for i, item := range items {
		res[i] = wire.QuickAccessItemToWire(&item)
	}

	return c.JSON(wire.QuickAccessItemListResponse{
		Success: true,
		Data:    res,
	})
}
