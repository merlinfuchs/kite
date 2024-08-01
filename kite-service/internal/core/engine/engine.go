package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type Engine struct {
	sync.RWMutex

	config       EngineConfig
	appStore     store.AppStore
	logStore     store.LogStore
	commandStore store.CommandStore

	lastUpdate time.Time
	apps       map[string]*App
}

func NewEngine(
	config EngineConfig,
	appStore store.AppStore,
	logStore store.LogStore,
	commandStore store.CommandStore,
) *Engine {
	return &Engine{
		config:       config,
		appStore:     appStore,
		logStore:     logStore,
		commandStore: commandStore,
		apps:         make(map[string]*App),
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
				if err := m.populateCommands(ctx); err != nil {
					slog.With("error", err).Error("failed to populate commands in engine")
				}
				if err := m.populateEvents(ctx); err != nil {
					slog.With("error", err).Error("failed to populate events in engine")
				}
			case <-deployTicker.C:
				m.deployCommands(ctx)
			}
		}
	}()
}

func (m *Engine) populateCommands(ctx context.Context) error {
	commandIDs, err := m.commandStore.EnabledCommandIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled command IDs: %w", err)
	}

	lastUpdate := m.lastUpdate
	m.lastUpdate = time.Now().UTC()

	commands, err := m.commandStore.EnabledCommandsUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	m.Lock()
	defer m.Unlock()

	for _, command := range commands {
		app, ok := m.apps[command.AppID]
		if !ok {
			app = NewApp(m.config, command.AppID, m.appStore, m.logStore, m.commandStore)
			m.apps[command.AppID] = app
		}

		app.AddCommand(command)
	}

	for _, app := range m.apps {
		app.RemoveDanglingCommands(commandIDs)
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

				slog.With("app_id", app.id).Info("Deploying commands for app")
				if err := app.DeployCommands(ctx); err != nil {
					slog.With("app_id", app.id).With("error", err).Error("failed to deploy commands")
				}
			}()
		}
	}
}

func (m *Engine) populateEvents(_ context.Context) error {
	return nil
}

func (e *Engine) HandleEvent(appID string, event gateway.Event) {
	e.RLock()
	app := e.apps[appID]
	e.RUnlock()

	if app != nil {
		app.HandleEvent(appID, event)
	}
}

type EngineConfig struct {
	MaxStackDepth int
	MaxOperations int
	MaxActions    int
}
