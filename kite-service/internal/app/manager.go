package app

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"time"

	disgatway "github.com/disgoorg/disgo/gateway"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.AppProvider = (*AppManager)(nil)

type EventHandlerFunc func(
	appID distype.Snowflake,
	shardID int,
	eventType disgatway.EventType,
	eventData io.Reader,
)

type AppManager struct {
	sync.Mutex

	apps map[distype.Snowflake]*App

	appStore      store.AppStore
	appUsageStore store.AppUsageStore
	Engine        *engine.Engine
}

func NewManager(
	appStore store.AppStore,
	appUsageStore store.AppUsageStore,
) *AppManager {
	return &AppManager{
		apps: make(map[distype.Snowflake]*App),

		appStore:      appStore,
		appUsageStore: appUsageStore,
	}
}

func (m *AppManager) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break
		case <-ticker.C:
			apps, err := m.appStore.GetAppsWithValidToken(ctx)
			if err != nil {
				slog.With("error", err).Error("Failed to get apps from store")
				continue
			}

			appIDs := make([]distype.Snowflake, 0, len(apps))

			for _, app := range apps {
				err := m.AddApp(ctx, app.ID, &app)
				if err != nil {
					slog.With("error", err).Error("Failed to add gateway")
				}
				appIDs = append(appIDs, app.ID)
			}

			m.CleanupApps(ctx, appIDs)
		}
	}
}

func (m *AppManager) Close(ctx context.Context) {
	m.Lock()
	defer m.Unlock()

	for _, g := range m.apps {
		g.Close(ctx)
	}
}

func (m *AppManager) AddApp(ctx context.Context, appID distype.Snowflake, info *model.App) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.apps[appID]; ok {
		return nil
	}

	slog.Info("Adding app", "client_id", appID)

	g, err := NewApp(appID, info, m.appUsageStore, m.Engine)
	if err != nil {
		return err
	}

	g.Open(ctx)

	m.apps[appID] = g
	return nil
}

func (m *AppManager) App(appID distype.Snowflake) *App {
	m.Lock()
	defer m.Unlock()

	return m.apps[appID]
}

func (m *AppManager) AppState(appID distype.Snowflake) (store.AppStateProvider, error) {
	m.Lock()
	defer m.Unlock()

	a, ok := m.apps[appID]
	if !ok {
		return nil, store.ErrNotFound
	}

	return a, nil
}

func (m *AppManager) RemoveApp(ctx context.Context, appID distype.Snowflake) {
	m.Lock()
	defer m.Unlock()

	if g, ok := m.apps[appID]; ok {
		g.Close(ctx)
		delete(m.apps, appID)
	}
}

func (m *AppManager) CleanupApps(ctx context.Context, appIDs []distype.Snowflake) {
	m.Lock()
	defer m.Unlock()

	for appID, g := range m.apps {
		found := false
		for _, id := range appIDs {
			if id == appID {
				found = true
				break
			}
		}

		if !found {
			g.Close(ctx)
			delete(m.apps, appID)
		}
	}
}
