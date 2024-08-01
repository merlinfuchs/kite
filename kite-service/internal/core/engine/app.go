package engine

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/internal/core/flow"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type App struct {
	sync.RWMutex

	id string

	config               EngineConfig
	appStore             store.AppStore
	logStore             store.LogStore
	commandStore         store.CommandStore
	hasUndeployedChanges bool

	commands map[string]*Command
	events   map[string]interface{}

	providers flow.FlowProviders
}

func NewApp(config EngineConfig, id string, appStore store.AppStore, logStore store.LogStore, commandStore store.CommandStore) *App {
	providers := flow.FlowProviders{
		Discord: NewDiscordProvider(id, appStore),
		Log:     NewLogProvider(id, logStore),
	}

	return &App{
		id:           id,
		config:       config,
		appStore:     appStore,
		logStore:     logStore,
		commandStore: commandStore,
		commands:     make(map[string]*Command),
		events:       make(map[string]interface{}),
		providers:    providers,
	}
}

func (a *App) AddCommand(cmd *model.Command) {
	a.Lock()
	defer a.Unlock()

	command, err := NewCommand(a.config, cmd, a.logStore, a.providers)
	if err != nil {
		slog.With("error", err).Error("failed to create command")
		return
	}

	a.commands[cmd.ID] = command

	if !cmd.LastDeployedAt.Valid || cmd.LastDeployedAt.Time.Before(cmd.UpdatedAt) {
		a.hasUndeployedChanges = true
	}
}

func (a *App) RemoveDanglingCommands(commandIDs []string) {
	a.Lock()
	defer a.Unlock()

	commandIDMap := make(map[string]struct{}, len(commandIDs))
	for _, commandID := range commandIDs {
		commandIDMap[commandID] = struct{}{}
	}

	for cmdID := range a.commands {
		if _, ok := commandIDMap[cmdID]; !ok {
			delete(a.commands, cmdID)
		}
	}
}

func (a *App) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := a.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     a.id,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", a.id).Error("Failed to create log entry from engine app")
	}
}

func (a *App) HandleEvent(appID string, event gateway.Event) {
	a.RLock()
	defer a.RUnlock()

	switch e := event.(type) {
	case *gateway.InteractionCreateEvent:
		switch d := e.Data.(type) {
		case *discord.CommandInteraction:
			for _, command := range a.commands {
				if command.cmd.Name == d.Name {
					command.HandleEvent(appID, event)
				}
			}
		}
	}
}
