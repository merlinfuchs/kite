package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type Engine struct {
	sync.RWMutex

	env *Env

	lastUpdate time.Time
	apps       map[string]*App
}

func NewEngine(
	env *Env,
) *Engine {
	return &Engine{
		env:  env,
		apps: make(map[string]*App),
	}
}

func (e *Engine) SetAppStateManager(appStateManager store.AppStateManager) {
	e.env.AppStateManager = appStateManager
}

func (m *Engine) Run(ctx context.Context) {
	updateTicker := time.NewTicker(1 * time.Second)
	deployTicker := time.NewTicker(60 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				updateTicker.Stop()
				return
			case <-updateTicker.C:
				lastUpdate := m.lastUpdate
				m.lastUpdate = time.Now().UTC()

				if err := m.populatePlugins(ctx, lastUpdate); err != nil {
					slog.Error(
						"Failed to populate plugins in engine",
						slog.String("error", err.Error()),
					)
				}
			case <-deployTicker.C:
				m.deployCommands(ctx)
			}
		}
	}()
}

func (m *Engine) populatePlugins(ctx context.Context, lastUpdate time.Time) error {
	pluginInstanceIDs, err := m.env.PluginInstanceStore.EnabledPluginInstanceIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled plugin instances: %w", err)
	}

	pluginInstances, err := m.env.PluginInstanceStore.EnabledPluginInstancesUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get plugin instances: %w", err)
	}

	m.Lock()
	defer m.Unlock()

	for _, pluginInstance := range pluginInstances {
		app, ok := m.apps[pluginInstance.AppID]
		if !ok {
			app = NewApp(
				pluginInstance.AppID,
				m.env,
			)
			m.apps[pluginInstance.AppID] = app
		}

		err := app.UpdatePlugin(ctx, pluginInstance)
		if err != nil {
			slog.Error(
				"Failed to update plugin",
				slog.String("plugin_id", pluginInstance.PluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	for _, app := range m.apps {
		app.RemoveDanglingPlugins(pluginInstanceIDs[app.id])
	}

	return nil
}

func (m *Engine) deployCommands(ctx context.Context) {
	m.Lock()
	defer m.Unlock()

	for _, a := range m.apps {
		appClient, err := m.env.AppStateManager.AppClient(ctx, a.id)
		if err != nil {
			slog.Error(
				"Failed to get app client",
				slog.String("app_id", a.id),
				slog.String("error", err.Error()),
			)
			continue
		}

		if a.commandsOutdated {
			go func() {
				ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()

				slog.Debug(
					"Deploying commands for app",
					slog.String("app_id", a.id),
				)
				if err := a.deployCommands(ctx, appClient); err != nil {
					slog.Error(
						"Failed to deploy commands",
						slog.String("app_id", a.id),
						slog.String("error", err.Error()),
					)
				}
			}()
		}
	}
}

func (e *Engine) HandleEvent(appID string, session *state.State, event gateway.Event) {
	e.RLock()
	app := e.apps[appID]
	e.RUnlock()

	if app != nil {
		app.HandleEvent(appID, session, event)
	}
}

type EngineConfig struct {
	MaxStackDepth int
	MaxOperations int
	MaxCredits    int
}
