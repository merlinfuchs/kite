package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type EventHandler interface {
	HandleEvent(appID string, session *state.State, event gateway.Event)
}

type GatewayManagerConfig struct {
	ClusterCount int
	ClusterIndex int
}

type GatewayManager struct {
	sync.Mutex

	config       GatewayManagerConfig
	appStore     store.AppStore
	logStore     store.LogStore
	planManager  *plan.PlanManager
	eventHandler EventHandler
	tokenCrypt   *util.SymmetricCrypt

	lastUpdate time.Time
	gateways   map[string]*Gateway
}

func NewGatewayManager(
	appStore store.AppStore,
	logStore store.LogStore,
	planManager *plan.PlanManager,
	eventHandler EventHandler,
	tokenCrypt *util.SymmetricCrypt,
	config GatewayManagerConfig,
) *GatewayManager {
	return &GatewayManager{
		config:       config,
		appStore:     appStore,
		logStore:     logStore,
		planManager:  planManager,
		eventHandler: eventHandler,
		tokenCrypt:   tokenCrypt,
		gateways:     make(map[string]*Gateway),
	}
}

func (m *GatewayManager) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		if err := m.populateGateways(ctx); err != nil {
			slog.With("error", err).Error("failed to populate gateways")
		}

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := m.populateGateways(ctx); err != nil {
					slog.With("error", err).Error("failed to populate gateways")
				}
			}
		}
	}()
}

func (m *GatewayManager) populateGateways(ctx context.Context) error {
	appIDs, err := m.appStore.EnabledAppIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled apps: %w", err)
	}

	lastUpdate := m.lastUpdate
	m.lastUpdate = time.Now().UTC()

	apps, err := m.appStore.EnabledAppsUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get apps updated since %s: %w", lastUpdate, err)
	}

	if err := m.removeDanglingGateways(ctx, appIDs); err != nil {
		return fmt.Errorf("failed to remove dangling apps: %w", err)
	}

	if len(apps) == 0 {
		return nil
	}

	filteredApps := make([]*model.App, 0, len(apps))
	for _, app := range apps {
		if util.CluserForKey(app.ID, m.config.ClusterCount) != m.config.ClusterIndex {
			continue
		}
		filteredApps = append(filteredApps, app)
	}

	slog.Info(
		"Populating gateways",
		slog.Int("total_apps", len(apps)),
		slog.Int("filtered_apps", len(filteredApps)),
		slog.Int("cluster_count", m.config.ClusterCount),
		slog.Int("cluster_index", m.config.ClusterIndex),
	)

	if len(filteredApps) != 0 {
		for _, app := range filteredApps {
			// Starting thousands of gateways at once can cause problems internally.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				if err := m.addGateway(ctx, app); err != nil {
					slog.Error(
						"Failed to add gateway",
						slog.String("app_id", app.ID),
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}

	return nil
}

func (m *GatewayManager) removeDanglingGateways(ctx context.Context, appIDs []string) error {
	m.Lock()
	defer m.Unlock()

	lookupMap := make(map[string]struct{}, len(appIDs))
	for _, id := range appIDs {
		lookupMap[id] = struct{}{}
	}

	var removed int
	for id, gateway := range m.gateways {
		if _, ok := lookupMap[id]; !ok {
			// Close should timeout after 5 seconds
			if err := gateway.Close(); err != nil {
				slog.Error(
					"Failed to close gateway",
					slog.String("app_id", id),
					slog.String("error", err.Error()),
				)
			}

			delete(m.gateways, id)
			removed++
		}
	}

	if removed != 0 {
		slog.Info("Removed dangling gateways", slog.Int("count", removed))
	}

	return nil
}

func (m *GatewayManager) addGateway(ctx context.Context, app *model.App) error {
	m.Lock()
	defer m.Unlock()

	if g, ok := m.gateways[app.ID]; ok {
		if g.session.GatewayIsAlive() {
			go g.Update(ctx, app)
			return nil
		}

		// Some times arikawa fails to keep the gateway alive, so we need to
		// re-add it.
		if err := g.Close(); err != nil {
			return fmt.Errorf("failed to close gateway: %w", err)
		}
		delete(m.gateways, app.ID)
	}

	g, err := NewGateway(app, m.logStore, m.appStore, m.planManager, m.eventHandler, m.tokenCrypt)
	if err != nil {
		return fmt.Errorf("failed to create gateway: %w", err)
	}

	m.gateways[app.ID] = g

	return nil
}

func (m *GatewayManager) AppState(ctx context.Context, appID string) (store.AppStateStore, error) {
	m.Lock()
	defer m.Unlock()

	g, ok := m.gateways[appID]
	if !ok {
		return nil, store.ErrNotFound
	}

	return g, nil
}

func (m *GatewayManager) AppClient(ctx context.Context, appID string) (*api.Client, error) {
	m.Lock()
	defer m.Unlock()

	g, ok := m.gateways[appID]
	if !ok {
		return nil, store.ErrNotFound
	}

	return g.session.Client, nil
}
