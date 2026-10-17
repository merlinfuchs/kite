package engine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type Command struct {
	cmd  *model.Command
	flow *flow.CompiledFlowNode
	env  Env
}

func NewCommand(
	cmd *model.Command,
	env Env,
) (*Command, error) {
	flow, err := flow.CompileCommand(cmd.FlowSource)
	if err != nil {
		slog.Error(
			"Failed to compile command flow",
			slog.String("app_id", cmd.AppID),
			slog.String("command_id", cmd.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to compile command flow: %w", err)
	}

	return &Command{
		cmd:  cmd,
		flow: flow,
		env:  env,
	}, nil
}

// commandType returns the type of the command. Commands without a compiled flow
// are slash commands.
func (c *Command) commandType() discord.CommandType {
	if c.flow == nil {
		return discord.ChatInputCommand
	}
	return c.flow.CommandType()
}

func (c *Command) HandleEvent(appID string, session *state.State, event gateway.Event) {
	links := entityLinks{
		CommandID: null.NewString(c.cmd.ID, true),
	}

	c.env.executeFlowEvent(
		context.Background(),
		c.cmd.AppID,
		c.flow,
		session,
		event,
		links,
		nil,
	)
}
