package engine

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

type App struct {
	sync.RWMutex

	id string

	env *Env

	commandsDeployedAt time.Time
	commandsOutdated   bool

	plugins map[string]*PluginInstance
}

func NewApp(
	id string,
	env *Env,
) *App {
	app := &App{
		id:      id,
		env:     env,
		plugins: make(map[string]*PluginInstance),
	}

	return app
}

func (a *App) UpdatePlugin(ctx context.Context, instance *model.PluginInstance) error {
	a.Lock()
	defer a.Unlock()

	pl, ok := a.plugins[instance.PluginID]
	if !ok {
		var err error
		pl, err = NewPluginInstance(a.id, instance, a.env)
		if err != nil {
			return fmt.Errorf("failed to create plugin instance: %w", err)
		}
	}

	discord, err := a.env.AppStateManager.AppClient(ctx, a.id)
	if err != nil {
		return fmt.Errorf("failed to get app client: %w", err)
	}

	err = pl.Update(ctx, instance, discord)
	if err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}

	if !a.commandsOutdated {
		a.commandsOutdated = !instance.CommandsDeployedAt.Valid || instance.CommandsDeployedAt.Time.Before(instance.UpdatedAt)
	}
	return nil
}

func (a *App) RemoveDanglingPlugins(pluginIDs []string) {
	a.Lock()
	defer a.Unlock()

	pluginIDMap := make(map[string]struct{}, len(pluginIDs))
	for _, pluginID := range pluginIDs {
		pluginIDMap[pluginID] = struct{}{}
	}

	for pluginID := range a.plugins {
		if _, ok := pluginIDMap[pluginID]; !ok {
			delete(a.plugins, pluginID)
		}
	}
}

func (a *App) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := a.env.LogStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     a.id,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", a.id).Error("Failed to create log entry from engine app")
	}
}

func (a *App) deployCommands(ctx context.Context, appClient *state.State) error {
	a.Lock()
	a.commandsOutdated = false

	deploymentStart := time.Now().UTC()

	var commands []api.CreateCommandData
	for _, plugin := range a.plugins {
		cmds := plugin.instance.Commands()
		for _, cmd := range cmds {
			commands = append(commands, cmd.Data)
		}
	}

	a.Unlock()

	app, err := a.env.AppStore.App(ctx, a.id)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	appId, err := strconv.ParseUint(app.DiscordID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse app ID: %w", err)
	}

	_, err = appClient.BulkOverwriteCommands(discord.AppID(appId), commands)
	if err != nil {
		go a.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to deploy commands: %v", err))
		return nil
	}

	err = a.env.CommandStore.UpdateCommandsLastDeployedAt(ctx, a.id, deploymentStart)
	if err != nil {
		return fmt.Errorf("failed to update last deployed at: %w", err)
	}

	go a.createLogEntry(model.LogLevelInfo, "Successfully deployed commands")
	return nil
}

func (a *App) HandleEvent(appID string, session *state.State, event gateway.Event) {
	a.RLock()
	defer a.RUnlock()

	for _, plugin := range a.plugins {
		// TODO: check if plugin is interested in this event
		plugin.HandleEvent(appID, session, event)
	}
}
