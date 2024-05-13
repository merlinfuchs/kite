package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/merlinfuchs/dismod/distype"
)

func (m *AppManager) populateRunningApps(ctx context.Context) error {
	apps, err := m.appStore.GetAppsWithValidToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get apps with valid token: %w", err)
	}

	m.lastUpdate = time.Now().UTC()

	appIDs := make([]distype.Snowflake, 0, len(apps))
	for _, app := range apps {
		err := m.AddApp(ctx, app.ID, &app)
		if err != nil {
			slog.With("error", err).With("app_id", app.ID).Error("Failed to add app gateway")
		}
		appIDs = append(appIDs, app.ID)
	}

	m.CleanupApps(ctx, appIDs)

	return nil
}

func (m *AppManager) updateRunningApps(ctx context.Context) error {
	apps, err := m.appStore.GetAppsWithValidToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get apps with valid token: %w", err)
	}

	lastUpdate := m.lastUpdate
	m.lastUpdate = time.Now().UTC()

	appIDs := make([]distype.Snowflake, 0, len(apps))
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)

		if app.UpdatedAt.After(lastUpdate) {
			m.RemoveApp(ctx, app.ID)

			err := m.AddApp(ctx, app.ID, &app)
			if err != nil {
				slog.With("error", err).With("app_id", app.ID).Error("Failed to add app gateway")
			}
		}
	}

	m.CleanupApps(ctx, appIDs)

	return nil
}
