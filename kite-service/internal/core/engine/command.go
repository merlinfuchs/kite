package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/template"
)

type Command struct {
	config     EngineConfig
	cmd        *model.Command
	flow       *flow.CompiledFlowNode
	appStore   store.AppStore
	logStore   store.LogStore
	httpClient *http.Client
}

func NewCommand(
	config EngineConfig,
	cmd *model.Command,
	appStore store.AppStore,
	logStore store.LogStore,
	httpClient *http.Client,
) (*Command, error) {
	flow, err := flow.CompileCommand(cmd.FlowSource)
	if err != nil {
		return nil, fmt.Errorf("failed to compile command flow: %w", err)
	}

	return &Command{
		config:     config,
		cmd:        cmd,
		flow:       flow,
		appStore:   appStore,
		logStore:   logStore,
		httpClient: httpClient,
	}, nil
}

func (c *Command) HandleEvent(appID string, session *state.State, event gateway.Event) {
	i, ok := event.(*gateway.InteractionCreateEvent)
	if !ok {
		return
	}

	providers := flow.FlowProviders{
		Discord: NewDiscordProvider(appID, c.appStore, session),
		Log:     NewLogProvider(appID, c.logStore),
		HTTP:    NewHTTPProvider(c.httpClient),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fCtx := flow.NewContext(ctx,
		&InteractionData{
			interaction: &i.InteractionEvent,
		},
		providers,
		flow.FlowContextLimits{
			MaxStackDepth: c.config.MaxStackDepth,
			MaxOperations: c.config.MaxOperations,
			MaxActions:    c.config.MaxActions,
		},
		template.NewContext(appID, 0),
	)

	if err := c.flow.Execute(fCtx); err != nil {
		go c.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to execute command flow: %v", err))
		slog.With("error", err).Error("Failed to execute command flow")
	}
}

func (a *App) DeployCommands(ctx context.Context) error {
	a.Lock()
	a.hasUndeployedChanges = false

	var lastUpdatedAt time.Time
	commands := make([]api.CreateCommandData, 0, len(a.commands))
	for _, command := range a.commands {
		cmd := command.cmd

		if cmd.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = cmd.UpdatedAt
		}

		node, err := flow.CompileCommand(cmd.FlowSource)
		if err != nil {
			go a.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to compile command flow: %v", err))
			return fmt.Errorf("failed to compile command flow: %w", err)
		}

		perms := node.CommandPermissions()

		commands = append(commands, api.CreateCommandData{
			Name:                     node.CommandName(),
			Description:              node.CommandDescription(),
			Options:                  node.CommandArguments(),
			DefaultMemberPermissions: &perms,
			NoDMPermission:           slices.Contains(node.CommandDisabledContexts(), flow.CommandContextTypeBotDM),
		})
	}

	a.Unlock()

	err := a.commandStore.UpdateCommandsLastDeployedAt(ctx, a.id, lastUpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update last deployed at: %w", err)
	}

	app, err := a.appStore.App(ctx, a.id)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	appId, err := strconv.ParseUint(app.DiscordID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse app ID: %w", err)
	}

	client := api.NewClient("Bot " + app.DiscordToken).WithContext(ctx)

	_, err = client.BulkOverwriteCommands(discord.AppID(appId), commands)
	if err != nil {
		go a.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to deploy commands: %v", err))
		return fmt.Errorf("failed to deploy commands: %w", err)
	}

	go a.createLogEntry(model.LogLevelInfo, "Successfully deployed commands")
	return nil
}

func (c *Command) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := c.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     c.cmd.AppID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", c.cmd.AppID).Error("Failed to create log entry from engine command")
	}
}
