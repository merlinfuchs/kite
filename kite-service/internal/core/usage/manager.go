package usage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

type UsageManager struct {
	appStore   store.AppStore
	usageStore store.UsageStore

	creditsPerMonth int
}

func NewUsageManager(appStore store.AppStore, usageStore store.UsageStore, creditsPerMonth int) *UsageManager {
	return &UsageManager{
		appStore:        appStore,
		usageStore:      usageStore,
		creditsPerMonth: creditsPerMonth,
	}
}

func (m *UsageManager) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := m.disableAppsWithNoCredits(ctx); err != nil {
					slog.Error(
						"Failed to disable apps with no credits",
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}()
}

func (m *UsageManager) disableAppsWithNoCredits(ctx context.Context) error {
	start, end := startAndEndOfMonth(time.Now().UTC())

	creditsUsed, err := m.usageStore.AllUsageCreditsUsedBetween(ctx, start, end)
	if err != nil {
		return fmt.Errorf("failed to get all usage credits used: %w", err)
	}

	for appID, creditsUsed := range creditsUsed {
		if creditsUsed >= m.creditsPerMonth {
			dCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			err := m.appStore.DisableApp(dCtx, store.AppDisableOpts{
				ID:             appID,
				DisabledReason: null.StringFrom("No credits remaining"),
				UpdatedAt:      time.Now().UTC(),
			})
			if err != nil {
				slog.Error(
					"Failed to disable app with no credits",
					slog.String("app_id", appID),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	return nil
}

func startAndEndOfMonth(t time.Time) (time.Time, time.Time) {
	year, month, _ := t.Date()
	start := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end
}
