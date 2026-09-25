package usage

import (
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/util"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type UsageHandler struct {
	usageStore store.UsageStore
}

func NewUsageHandler(usageStore store.UsageStore) *UsageHandler {
	return &UsageHandler{
		usageStore: usageStore,
	}
}

func (h *UsageHandler) HandleUsageCreditsGet(c *handler.Context) (*wire.UsageCreditsGetResponse, error) {
	start, end := util.StartAndEndOfMonth(time.Now().UTC())

	creditsUsed, err := h.usageStore.UsageCreditsUsedBetween(c.Context(), c.App.ID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get total credits: %w", err)
	}

	return &wire.UsageCreditsGetResponse{
		CreditsUsed: creditsUsed,
	}, nil
}

func (h *UsageHandler) HandleUsageByDayList(c *handler.Context) (*wire.UsageByDayListResponse, error) {
	start, end := util.StartAndEndOfMonth(time.Now().UTC())

	entries, err := h.usageStore.UsageCreditsUsedByDayBetween(c.Context(), c.App.ID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by day: %w", err)
	}

	res := make(wire.UsageByDayListResponse, 0, len(entries))
	for _, entry := range entries {
		res = append(res, &wire.UsageByDayEntry{
			Date:        entry.Date,
			CreditsUsed: entry.CreditsUsed,
		})
	}

	return &res, nil
}

func (h *UsageHandler) HandleUsageByTypeList(c *handler.Context) (*wire.UsageByTypeListResponse, error) {
	start, end := util.StartAndEndOfMonth(time.Now().UTC())

	entries, err := h.usageStore.UsageCreditsUsedByTypeBetween(c.Context(), c.App.ID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by type: %w", err)
	}

	res := make(wire.UsageByTypeListResponse, 0, len(entries))
	for _, entry := range entries {
		res = append(res, &wire.UsageByTypeEntry{
			Type:        string(entry.Type),
			CreditsUsed: entry.CreditsUsed,
		})
	}

	return &res, nil
}
