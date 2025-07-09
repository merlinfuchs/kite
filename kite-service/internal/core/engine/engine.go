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

func (e *Engine) Run(ctx context.Context) {
	updateTicker := time.NewTicker(1 * time.Second)
	removeTicker := time.NewTicker(60 * time.Second)
	deployTicker := time.NewTicker(60 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				updateTicker.Stop()
				return
			case <-updateTicker.C:
				lastUpdate := e.lastUpdate
				e.lastUpdate = time.Now().UTC()

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
			case <-deployTicker.C:
				e.deployCommands(ctx)
			}
		}
	}()
}

func (e *Engine) populateCommands(ctx context.Context, lastUpdate time.Time) error {
	commands, err := e.stores.CommandStore.EnabledCommandsUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	lockStart := time.Now()
	e.Lock()
	defer e.Unlock()
	lockDiff := time.Since(lockStart)
	if lockDiff > 250*time.Millisecond {
		slog.Warn(
			"Locking engine for commands took too long",
			slog.String("lock_duration", lockDiff.String()),
		)
	}

	for _, command := range commands {
		app, ok := e.apps[command.AppID]
		if !ok {
			app = NewApp(
				command.AppID,
				e.stores,
			)
			e.apps[command.AppID] = app
		}

		app.AddCommand(command)
	}

	return nil
}

func (e *Engine) removeDanglingCommands(ctx context.Context) error {
	commandIDs, err := e.stores.CommandStore.EnabledCommandIDs(ctx)
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
	listeners, err := e.stores.EventListenerStore.EnabledEventListenersUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get event listeners: %w", err)
	}

	e.Lock()
	defer e.Unlock()

	for _, listener := range listeners {
		app, ok := e.apps[listener.AppID]
		if !ok {
			app = NewApp(
				listener.AppID,
				e.stores,
			)
			e.apps[listener.AppID] = app
		}

		app.AddEventListener(listener)
	}

	return nil
}

func (e *Engine) removeDanglingEventListeners(ctx context.Context) error {
	listenerIDs, err := e.stores.EventListenerStore.EnabledEventListenerIDs(ctx)
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

func (e *Engine) deployCommands(ctx context.Context) {
	e.Lock()
	defer e.Unlock()

	for _, app := range e.apps {
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

// HandleEvent blocks until the event is handled by the corresponding app.
func (e *Engine) HandleEvent(appID string, session *state.State, event gateway.Event) {
	lockStart := time.Now()
	e.RLock()
	app := e.apps[appID]
	e.RUnlock()
	lockDiff := time.Since(lockStart)
	if lockDiff > 100*time.Millisecond {
		slog.Warn(
			"Locking engine for event took too long",
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
}
