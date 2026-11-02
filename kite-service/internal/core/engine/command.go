package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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

// autocompleteTimeout is how long the blocks after an autocomplete output can
// run. Discord only waits 3 seconds for the choices.
const autocompleteTimeout = 2500 * time.Millisecond

// HandleAutocomplete shows the choices for the argument the user is typing in.
// Arguments with blocks connected to their autocomplete output run them, the
// others show those of their choices that match what was typed, without
// using credits.
func (c *Command) HandleAutocomplete(appID string, session *state.State, event *gateway.InteractionCreateEvent, focused discord.AutocompleteOption) {
	arg := c.flow.CommandArgument(focused.Name)
	if arg == nil || !arg.HasAutocompleteFlow() {
		err := session.RespondInteraction(event.ID, event.Token, arg.StaticAutocompleteResponse(focused))
		if err != nil {
			slog.Warn(
				"Failed to respond to autocomplete interaction",
				slog.String("app_id", appID),
				slog.String("command_id", c.cmd.ID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), autocompleteTimeout)
	defer cancel()

	c.env.executeFlowAutocomplete(
		ctx,
		c.cmd.AppID,
		arg,
		session,
		event,
		entityLinks{CommandID: null.NewString(c.cmd.ID, true)},
	)
}
