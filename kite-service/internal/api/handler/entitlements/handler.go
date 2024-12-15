package entitlements

import (
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type EntitlementsHandler struct {
	usageStore      store.UsageStore
	creditsPerMonth int
}

func NewEntitlementsHandler(usageStore store.UsageStore, creditsPerMonth int) *EntitlementsHandler {
	return &EntitlementsHandler{
		usageStore:      usageStore,
		creditsPerMonth: creditsPerMonth,
	}
}

func (h *EntitlementsHandler) HandleEntitlementsCreditsGet(c *handler.Context) (*wire.EntitlementsCreditsGetResponse, error) {
	start, end := startAndEndOfMonth(time.Now())

	creditsUsed, err := h.usageStore.UsageCreditsUsedBetween(c.Context(), c.App.ID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get total credits: %w", err)
	}

	return &wire.EntitlementsCreditsGetResponse{
		TotalCredits: h.creditsPerMonth,
		CreditsUsed:  creditsUsed,
	}, nil
}

func startAndEndOfMonth(t time.Time) (time.Time, time.Time) {
	year, month, _ := t.Date()
	start := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end
}
