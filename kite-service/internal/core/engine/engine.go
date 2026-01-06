package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type Engine struct {
	sync.RWMutex

	env Env

	lastUpdate time.Time
	apps       map[string]*App
}

func NewEngine(
	env Env,
) *Engine {
	return &Engine{
		env:  env,
		apps: make(map[string]*App),
	}
}

func (e *Engine) Run(ctx context.Context) {
	go func() {
		updateTicker := time.NewTicker(1 * time.Second)
		defer updateTicker.Stop()

		removeTicker := time.NewTicker(60 * time.Second)
		defer removeTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-updateTicker.C:
				lastUpdate := e.lastUpdate
				e.lastUpdate = time.Now().UTC()

				if err := e.populatePlugins(ctx, lastUpdate); err != nil {
					slog.Error(
						"Failed to populate plugins in engine",
						slog.String("error", err.Error()),
					)
				}
				if err := e.populateCommands(ctx, lastUpdate); err != nil {
					slog.Error(
						"Failed to populate commands in engine",
						slog.String("error", err.Error()),
					)
				}
				if err := e.populateEventListeners(ctx, lastUpdate); err != nil {
					slog.Error(
						"Failed to populate event listeners in engine",
						slog.String("error", err.Error()),
					)
				}
			case <-removeTicker.C:
				if err := e.removeDanglingPlugins(ctx); err != nil {
					slog.Error(
						"Failed to remove dangling plugins in engine",
						slog.String("error", err.Error()),
					)
				}
				if err := e.removeDanglingCommands(ctx); err != nil {
					slog.Error(
						"Failed to remove dangling commands in engine",
						slog.String("error", err.Error()),
					)
				}
				if err := e.removeDanglingEventListeners(ctx); err != nil {
					slog.Error(
						"Failed to remove dangling event listeners in engine",
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}()
}

func (e *Engine) populatePlugins(ctx context.Context, lastUpdate time.Time) error {
	pluginInstances, err := e.env.PluginInstanceStore.EnabledPluginInstancesUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get plugin instances: %w", err)
	}

	lockStart := time.Now()
	e.Lock()
	defer e.Unlock()
	lockDiff := time.Since(lockStart)
	if lockDiff > 5*time.Second {
		slog.Warn(
			"Locking engine for plugins took too long",
			slog.String("lock_duration", lockDiff.String()),
		)
	}

	for _, pluginInstance := range pluginInstances {
		if util.CluserForKey(pluginInstance.AppID, e.env.Config.ClusterCount) != e.env.Config.ClusterIndex {
			continue
		}

		app, ok := e.apps[pluginInstance.AppID]
		if !ok {
			app = NewApp(
				pluginInstance.AppID,
				e.env,
			)
			e.apps[pluginInstance.AppID] = app
		}

		app.AddPluginInstance(pluginInstance)
	}

	return nil
}

func (e *Engine) removeDanglingPlugins(ctx context.Context) error {
	pluginInstanceIDs, err := e.env.PluginInstanceStore.EnabledPluginInstanceIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled plugin instance IDs: %w", err)
	}

	e.RLock()
	defer e.RUnlock()

	for _, app := range e.apps {
		app.RemoveDanglingPluginInstances(pluginInstanceIDs)
	}

	return nil
}

func (e *Engine) populateCommands(ctx context.Context, lastUpdate time.Time) error {
	commands, err := e.env.CommandStore.EnabledCommandsUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	lockStart := time.Now()
	e.Lock()
	defer e.Unlock()
	lockDiff := time.Since(lockStart)
	if lockDiff > 5*time.Second {
		slog.Warn(
			"Locking engine for commands took too long",
			slog.String("lock_duration", lockDiff.String()),
		)
	}

	for _, command := range commands {
		if util.CluserForKey(command.AppID, e.env.Config.ClusterCount) != e.env.Config.ClusterIndex {
			continue
		}

		app, ok := e.apps[command.AppID]
		if !ok {
			app = NewApp(
				command.AppID,
				e.env,
			)
			e.apps[command.AppID] = app
		}

		app.AddCommand(command)
	}

	return nil
}

func (e *Engine) removeDanglingCommands(ctx context.Context) error {
	commandIDs, err := e.env.CommandStore.EnabledCommandIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled command IDs: %w", err)
	}

	e.RLock()
	defer e.RUnlock()

	for _, app := range e.apps {
		app.RemoveDanglingCommands(commandIDs)
	}

	return nil
}

func (e *Engine) populateEventListeners(ctx context.Context, lastUpdate time.Time) error {
	listeners, err := e.env.EventListenerStore.EnabledEventListenersUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get event listeners: %w", err)
	}

	e.Lock()
	defer e.Unlock()

	for _, listener := range listeners {
		if util.CluserForKey(listener.AppID, e.env.Config.ClusterCount) != e.env.Config.ClusterIndex {
			continue
		}

		app, ok := e.apps[listener.AppID]
		if !ok {
			app = NewApp(
				listener.AppID,
				e.env,
			)
			e.apps[listener.AppID] = app
		}

		app.AddEventListener(listener)
	}

	return nil
}

func (e *Engine) removeDanglingEventListeners(ctx context.Context) error {
	listenerIDs, err := e.env.EventListenerStore.EnabledEventListenerIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled event listener IDs: %w", err)
	}

	e.RLock()
	defer e.RUnlock()

	for _, app := range e.apps {
		app.RemoveDanglingEventListeners(listenerIDs)
	}

	return nil
}

// HandleEvent blocks until the event is handled by the corresponding app.
func (e *Engine) HandleEvent(appID string, session *state.State, event gateway.Event) {
	lockStart := time.Now()
	e.RLock()
	app := e.apps[appID]
	e.RUnlock()
	lockDiff := time.Since(lockStart)
	if lockDiff > 500*time.Millisecond {
		slog.Warn(
			"Locking engine for handling event took too long",
			slog.String("app_id", appID),
			slog.String("lock_duration", lockDiff.String()),
		)
	}

	if app != nil {
		app.HandleEvent(appID, session, event)
	}
}

type EngineConfig struct {
	MaxStackDepth int
	MaxOperations int
	MaxCredits    int
	ClusterCount  int
	ClusterIndex  int
}
