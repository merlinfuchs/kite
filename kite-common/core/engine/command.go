package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-common/core/flow"
	"github.com/kitecloud/kite/kite-common/model"
	"github.com/kitecloud/kite/kite-common/store"
)

type Command struct {
	config    EngineConfig
	cmd       *model.Command
	flow      *flow.CompiledFlowNode
	providers flow.FlowProviders
	logStore  store.LogStore
}

func NewCommand(config EngineConfig, cmd *model.Command, logStore store.LogStore, providers flow.FlowProviders) (*Command, error) {
	flow, err := flow.CompileCommand(cmd.FlowSource)
	if err != nil {
		return nil, fmt.Errorf("failed to compile command flow: %w", err)
	}

	return &Command{
		config:    config,
		cmd:       cmd,
		flow:      flow,
		providers: providers,
		logStore:  logStore,
	}, nil
}

func (c *Command) HandleEvent(appID string, event gateway.Event) {
	i, ok := event.(*gateway.InteractionCreateEvent)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fCtx := flow.NewContext(ctx,
		&InteractionData{
			interaction: &i.InteractionEvent,
		},
		c.providers,
		flow.FlowContextLimits{
			MaxStackDepth: c.config.MaxStackDepth,
			MaxOperations: c.config.MaxOperations,
			MaxActions:    c.config.MaxActions,
		},
	)

	if err := c.flow.Execute(fCtx); err != nil {
		go c.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to execute command flow: %v", err))
		slog.With("error", err).Error("Failed to execute command flow")
	}
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
