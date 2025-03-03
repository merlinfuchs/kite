package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

type Engine struct {
	sync.RWMutex

	stores Env

	lastUpdate time.Time
	apps       map[string]*App
}

func NewEngine(
	stores Env,
) *Engine {
	return &Engine{
		stores: stores,
		apps:   make(map[string]*App),
	}
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

				if err := m.populateCommands(ctx, lastUpdate); err != nil {
					slog.Error(
						"Failed to populate commands in engine",
						slog.String("error", err.Error()),
					)
				}
				if err := m.populateEventListeners(ctx, lastUpdate); err != nil {
					slog.Error(
						"Failed to populate event listeners in engine",
						slog.String("error", err.Error()),
					)
				}
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

func (m *Engine) populateCommands(ctx context.Context, lastUpdate time.Time) error {
	commandIDs, err := m.stores.CommandStore.EnabledCommandIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled command IDs: %w", err)
	}

	commands, err := m.stores.CommandStore.EnabledCommandsUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	m.Lock()
	defer m.Unlock()

	for _, command := range commands {
		app, ok := m.apps[command.AppID]
		if !ok {
			app = NewApp(
				command.AppID,
				m.stores,
			)
			m.apps[command.AppID] = app
		}

		app.AddCommand(command)
	}

	for _, app := range m.apps {
		app.RemoveDanglingCommands(commandIDs)
	}

	return nil
}

func (m *Engine) populateEventListeners(ctx context.Context, lastUpdate time.Time) error {
	listenerIDs, err := m.stores.EventListenerStore.EnabledEventListenerIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled event listener IDs: %w", err)
	}

	listeners, err := m.stores.EventListenerStore.EnabledEventListenersUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get event listeners: %w", err)
	}

	m.Lock()
	defer m.Unlock()

	for _, listener := range listeners {
		app, ok := m.apps[listener.AppID]
		if !ok {
			app = NewApp(
				listener.AppID,
				m.stores,
			)
			m.apps[listener.AppID] = app
		}

		app.AddEventListener(listener)
	}

	for _, app := range m.apps {
		app.RemoveDanglingEventListeners(listenerIDs)
	}

	return nil
}

func (m *Engine) populatePlugins(ctx context.Context, lastUpdate time.Time) error {
	pluginInstanceIDs, err := m.stores.PluginInstanceStore.EnabledPluginInstanceIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled plugin instances: %w", err)
	}

	pluginInstances, err := m.stores.PluginInstanceStore.EnabledPluginInstancesUpdatedSince(ctx, lastUpdate)
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
				m.stores,
			)
			m.apps[pluginInstance.AppID] = app
		}

		app.AddPlugin(pluginInstance)
	}

	for _, app := range m.apps {
		app.RemoveDanglingPlugins(pluginInstanceIDs[app.id])
	}

	return nil
}

func (m *Engine) deployCommands(ctx context.Context) {
	m.Lock()
	defer m.Unlock()

	for _, app := range m.apps {
		if app.hasUndeployedChanges {
			go func() {
				ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()

				slog.Debug(
					"Deploying commands for app",
					slog.String("app_id", app.id),
				)
				if err := app.DeployCommands(ctx); err != nil {
					slog.Error(
						"Failed to deploy commands",
						slog.String("app_id", app.id),
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
