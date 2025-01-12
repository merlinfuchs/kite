package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/sashabaranov/go-openai"
)

type Engine struct {
	sync.RWMutex

	config               EngineConfig
	appStore             store.AppStore
	logStore             store.LogStore
	usageStore           store.UsageStore
	messageStore         store.MessageStore
	messageInstanceStore store.MessageInstanceStore
	commandStore         store.CommandStore
	eventListenerStore   store.EventListenerStore
	variableValueStore   store.VariableValueStore
	suspendPointStore    store.SuspendPointStore
	httpClient           *http.Client
	openaiClient         *openai.Client

	lastUpdate time.Time
	apps       map[string]*App
}

func NewEngine(
	config EngineConfig,
	appStore store.AppStore,
	logStore store.LogStore,
	usageStore store.UsageStore,
	messageStore store.MessageStore,
	messageInstanceStore store.MessageInstanceStore,
	commandStore store.CommandStore,
	eventListenerStore store.EventListenerStore,
	variableValueStore store.VariableValueStore,
	suspendPointStore store.SuspendPointStore,
	httpClient *http.Client,
	openaiClient *openai.Client,
) *Engine {
	return &Engine{
		config:               config,
		appStore:             appStore,
		logStore:             logStore,
		usageStore:           usageStore,
		messageStore:         messageStore,
		messageInstanceStore: messageInstanceStore,
		httpClient:           httpClient,
		commandStore:         commandStore,
		eventListenerStore:   eventListenerStore,
		variableValueStore:   variableValueStore,
		suspendPointStore:    suspendPointStore,
		openaiClient:         openaiClient,
		apps:                 make(map[string]*App),
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
			case <-deployTicker.C:
				m.deployCommands(ctx)
			}
		}
	}()
}

func (m *Engine) populateCommands(ctx context.Context, lastUpdate time.Time) error {
	commandIDs, err := m.commandStore.EnabledCommandIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled command IDs: %w", err)
	}

	commands, err := m.commandStore.EnabledCommandsUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	m.Lock()
	defer m.Unlock()

	for _, command := range commands {
		app, ok := m.apps[command.AppID]
		if !ok {
			app = NewApp(
				m.config,
				command.AppID,
				m.appStore,
				m.logStore,
				m.usageStore,
				m.messageStore,
				m.messageInstanceStore,
				m.commandStore,
				m.variableValueStore,
				m.suspendPointStore,
				m.httpClient,
				m.openaiClient,
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
	listenerIDs, err := m.eventListenerStore.EnabledEventListenerIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled event listener IDs: %w", err)
	}

	listeners, err := m.eventListenerStore.EnabledEventListenersUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get event listeners: %w", err)
	}

	m.Lock()
	defer m.Unlock()

	for _, listener := range listeners {
		app, ok := m.apps[listener.AppID]
		if !ok {
			app = NewApp(
				m.config,
				listener.AppID,
				m.appStore,
				m.logStore,
				m.usageStore,
				m.messageStore,
				m.messageInstanceStore,
				m.commandStore,
				m.variableValueStore,
				m.suspendPointStore,
				m.httpClient,
				m.openaiClient,
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
