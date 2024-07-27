package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-common/core/flow"
	"github.com/kitecloud/kite/kite-common/model"
)

type Command struct {
	cmd       *model.Command
	flow      *flow.CompiledFlowNode
	providers flow.FlowProviders
}

func NewCommand(cmd *model.Command, providers flow.FlowProviders) (*Command, error) {
	flow, err := flow.CompileCommand(cmd.FlowSource)
	if err != nil {
		return nil, fmt.Errorf("failed to compile command flow: %w", err)
	}

	return &Command{
		cmd:       cmd,
		flow:      flow,
		providers: providers,
	}, nil
}

func (a *Command) HandleEvent(appID string, event gateway.Event) {
	i, ok := event.(*gateway.InteractionCreateEvent)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := flow.NewContext(ctx,
		&InteractionData{
			interaction: &i.InteractionEvent,
		},
		a.providers,
		flow.FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 100,
			MaxActions:    10,
		},
	)

	if err := a.flow.Execute(c); err != nil {
		slog.With("error", err).Error("Failed to execute command flow")
	}
}
