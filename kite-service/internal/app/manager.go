package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sync"
	"time"

	disgatway "github.com/disgoorg/disgo/gateway"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
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

	apps       map[distype.Snowflake]*App
	lastUpdate time.Time

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

func (m *AppManager) Start(ctx context.Context) error {
	err := m.populateRunningApps(ctx)
	if err != nil {
		return fmt.Errorf("error populating running apps: %w", err)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				break
			case <-ticker.C:
				err := m.updateRunningApps(ctx)
				if err != nil {
					slog.With(logattr.Error(err)).Error("Error updating engine deployments")
				}
			}
		}
	}()

	return nil
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
		if slices.Contains(appIDs, appID) {
			continue
		}

		g.Close(ctx)
		delete(m.apps, appID)
	}
}
