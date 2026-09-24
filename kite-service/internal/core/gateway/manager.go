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
	"github.com/kitecloud/kite/kite-service/internal/metrics"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type EventHandler interface {
	HandleEvent(appID string, session *state.State, event gateway.Event)
}

type GatewayManagerConfig struct {
	ClusterCount int
	ClusterIndex int

	PopulateInterval       time.Duration
	RemoveDanglingInterval time.Duration
	PopulateOverlap        time.Duration
	StartInterval          time.Duration
}

// defaultStartInterval paces gateway starts when none is configured.
//
// This bounds the rate of new TLS handshakes, not anything Discord enforces:
// each app has its own bot token and so its own identify budget. Each cluster
// paces through its own share of apps independently, so the fleet-wide rate is
// this multiplied by the cluster count.
//
// Tune via gateway.start_interval. The trade is cold-start time against
// outbound connection rate: at ~9k apps per cluster, 100ms means roughly 15
// minutes to initiate them all, 10ms means about 90 seconds.
const defaultStartInterval = 100 * time.Millisecond

type GatewayManager struct {
	sync.Mutex

	config         GatewayManagerConfig
	appStore       store.AppStore
	logStore       store.LogStore
	planManager    *plan.PlanManager
	eventHandler   EventHandler
	tokenCrypt     *util.SymmetricCrypt
	pluginRegistry *plugin.Registry

	lastUpdate time.Time
	gateways   map[string]*Gateway
}

func NewGatewayManager(
	appStore store.AppStore,
	logStore store.LogStore,
	planManager *plan.PlanManager,
	eventHandler EventHandler,
	tokenCrypt *util.SymmetricCrypt,
	pluginRegistry *plugin.Registry,
	config GatewayManagerConfig,
) *GatewayManager {
	return &GatewayManager{
		config:         config,
		appStore:       appStore,
		logStore:       logStore,
		planManager:    planManager,
		eventHandler:   eventHandler,
		tokenCrypt:     tokenCrypt,
		pluginRegistry: pluginRegistry,
		gateways:       make(map[string]*Gateway),
	}
}

func (m *GatewayManager) Run(ctx context.Context) {
	populateInterval := util.IntervalOrDefault(m.config.PopulateInterval, 10*time.Second)
	removeDanglingInterval := util.IntervalOrDefault(m.config.RemoveDanglingInterval, 60*time.Second)

	go func() {
		populateTicker := time.NewTicker(populateInterval)
		defer populateTicker.Stop()

		// The full scan of enabled app IDs is by far the more expensive of the
		// two jobs, so it runs on its own slower ticker. Apps that get
		// disabled are still picked up promptly by the cheap query in
		// populateGateways; this only exists to catch hard deletes.
		removeDanglingTicker := time.NewTicker(removeDanglingInterval)
		defer removeDanglingTicker.Stop()

		if err := m.populateGateways(ctx); err != nil {
			slog.With("error", err).Error("failed to populate gateways")
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-populateTicker.C:
				if err := m.populateGateways(ctx); err != nil {
					slog.With("error", err).Error("failed to populate gateways")
				}
			case <-removeDanglingTicker.C:
				if err := m.removeDeletedGateways(ctx); err != nil {
					slog.With("error", err).Error("failed to remove deleted gateways")
				}
			}
		}
	}()
}

// populateGateways starts gateways for new or changed apps and stops gateways
// for apps that were disabled.
//
// The poll cursor is captured before the queries run and only advanced once
// they all succeed. Advancing it beforehand (or after a failure) silently
// drops every app updated while the queries were in flight, because the next
// poll would already consider that window covered.
func (m *GatewayManager) populateGateways(ctx context.Context) error {
	tickStart := time.Now().UTC()
	lastUpdate := m.lastUpdate

	queryStart := time.Now()
	apps, err := m.appStore.EnabledAppsUpdatedSince(ctx, lastUpdate)
	metrics.ObservePoll("gateway_enabled_apps_updated", queryStart)
	if err != nil {
		return fmt.Errorf("failed to get apps updated since %s: %w", lastUpdate, err)
	}

	queryStart = time.Now()
	disabledAppIDs, err := m.appStore.DisabledAppIDsUpdatedSince(ctx, lastUpdate)
	metrics.ObservePoll("gateway_disabled_apps_updated", queryStart)
	if err != nil {
		return fmt.Errorf("failed to get disabled apps updated since %s: %w", lastUpdate, err)
	}

	queryStart = time.Now()
	changedReqAppIDs, err := m.appStore.AppIDsWithGatewayRequirementsChangedSince(ctx, lastUpdate)
	metrics.ObservePoll("gateway_requirements_changed", queryStart)
	if err != nil {
		return fmt.Errorf("failed to get apps with changed gateway requirements since %s: %w", lastUpdate, err)
	}

	m.removeGateways(disabledAppIDs)

	// Must stay ahead of startGateways. On the first poll the cursor is zero,
	// so changedReqAppIDs contains every app that has any listener or plugin
	// instance -- tens of thousands. Running before any gateway exists makes
	// those all cheap map misses; running after would instead mean one
	// database query and one Discord round trip each, serially.
	//
	// Nothing is lost by refreshing first: gateways started below compute
	// their intents from the same rows before they connect.
	m.refreshIntents(ctx, changedReqAppIDs)

	filteredApps := make([]*model.App, 0, len(apps))
	for _, app := range apps {
		if util.CluserForKey(app.ID, m.config.ClusterCount) != m.config.ClusterIndex {
			continue
		}
		filteredApps = append(filteredApps, app)
	}

	if len(filteredApps) != 0 {
		slog.Info(
			"Populating gateways",
			slog.Int("total_apps", len(apps)),
			slog.Int("filtered_apps", len(filteredApps)),
			slog.Int("cluster_count", m.config.ClusterCount),
			slog.Int("cluster_index", m.config.ClusterIndex),
		)

		if err := m.startGateways(ctx, filteredApps); err != nil {
			return err
		}
	}

	// Rewind by the overlap so apps updated during the queries, or committed
	// out of timestamp order, are still picked up next time. Adding a gateway
	// that already exists is a no-op update, so re-reading rows is harmless.
	m.lastUpdate = tickStart.Add(-m.config.PopulateOverlap)

	return nil
}

// startGateways brings up gateways for the given apps, pacing the starts.
//
// Pacing rather than bounding concurrency is deliberate: addGateway returns as
// soon as the session is constructed and hands the actual connect off to a
// goroutine, so a concurrency limit around addGateway would bound almost
// nothing. Spacing the initiations is what actually keeps a cold start from
// opening tens of thousands of TLS handshakes at once.
//
// Every app has its own bot token and therefore its own identify budget, so
// this protects this process rather than satisfying a Discord rate limit.
func (m *GatewayManager) startGateways(ctx context.Context, apps []*model.App) error {
	ticker := time.NewTicker(util.IntervalOrDefault(m.config.StartInterval, defaultStartInterval))
	defer ticker.Stop()

	for _, app := range apps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		if err := m.addGateway(ctx, app); err != nil {
			slog.Error(
				"Failed to add gateway",
				slog.String("app_id", app.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

// refreshIntents recomputes intents for apps whose event listeners or plugin
// instances changed, reconnecting those whose required intent set moved.
//
// Each refresh costs two round trips (requirements plus the app's flags) and
// only runs for apps that actually changed, which in steady state is none.
func (m *GatewayManager) refreshIntents(ctx context.Context, appIDs []string) {
	for _, id := range appIDs {
		m.Lock()
		gateway, ok := m.gateways[id]
		m.Unlock()

		if !ok {
			continue
		}

		gateway.RefreshIntents(ctx)
	}
}

// closeGatewayLocked closes and deregisters the gateway for an app, reporting
// whether one was actually registered. Callers must hold the manager lock; the
// close itself runs in the background because it can block for seconds.
func (m *GatewayManager) closeGatewayLocked(appID string) bool {
	gateway, ok := m.gateways[appID]
	if !ok {
		return false
	}

	go func() {
		if err := gateway.Close(); err != nil {
			slog.Error(
				"Failed to close gateway",
				slog.String("app_id", appID),
				slog.String("error", err.Error()),
			)
		}
	}()

	delete(m.gateways, appID)
	metrics.GatewayConnections.Add(-1)

	return true
}

// removeGateways stops the gateways for the given app IDs, if this process
// owns them.
func (m *GatewayManager) removeGateways(appIDs []string) {
	if len(appIDs) == 0 {
		return
	}

	m.Lock()
	defer m.Unlock()

	var removed int
	for _, id := range appIDs {
		if m.closeGatewayLocked(id) {
			removed++
		}
	}

	if removed != 0 {
		slog.Info("Removed gateways for disabled apps", slog.Int("count", removed))
	}
}

// removeDeletedGateways scans every enabled app ID to find gateways whose app
// no longer exists. This is the expensive path, so it runs on its own slower
// ticker; disabled apps are handled by the cheaper query in populateGateways.
func (m *GatewayManager) removeDeletedGateways(ctx context.Context) error {
	queryStart := time.Now()
	appIDs, err := m.appStore.EnabledAppIDs(ctx)
	metrics.ObservePoll("gateway_enabled_app_ids", queryStart)
	if err != nil {
		return fmt.Errorf("failed to get enabled apps: %w", err)
	}

	return m.removeDanglingGateways(ctx, appIDs)
}

func (m *GatewayManager) removeDanglingGateways(ctx context.Context, appIDs []string) error {
	m.Lock()
	defer m.Unlock()

	lookupMap := util.IDSet(appIDs)

	var removed int
	for id := range m.gateways {
		if _, ok := lookupMap[id]; !ok {
			m.closeGatewayLocked(id)
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

		go func() {
			// Some times arikawa fails to keep the gateway alive, so we need to
			// re-add it.
			if err := g.Close(); err != nil {
				slog.Error("Failed to close gateway", slog.String("app_id", app.ID), slog.String("error", err.Error()))
			}
		}()
		delete(m.gateways, app.ID)
		metrics.GatewayConnections.Add(-1)
	}

	g, err := NewGateway(app, m.logStore, m.appStore, m.planManager, m.eventHandler, m.tokenCrypt, m.pluginRegistry)
	if err != nil {
		return fmt.Errorf("failed to create gateway: %w", err)
	}

	m.gateways[app.ID] = g
	metrics.GatewayConnections.Add(1)

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
