package flow

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
)

type FlowContext struct {
	context.Context
	FlowProviders
	FlowContextLimits
	FlowContextState

	Data    FlowContextData
	EvalCtx eval.Context
	Cancel  context.CancelFunc
}

func NewContext(
	ctx context.Context,
	timeout time.Duration,
	data FlowContextData,
	providers FlowProviders,
	limits FlowContextLimits,
	evalCtx eval.Context,
	state *FlowContextState,
) *FlowContext {
	ctx, cancel := context.WithTimeout(ctx, timeout)

	if state == nil {
		state = NewFlowContextState()
	}

	nodeEvalEnv := &nodeEvalEnv{
		state: state,
	}

	evalCtx.Env["node"] = nodeEvalEnv.GetNode
	evalCtx.Env["result"] = func(id string) (any, error) {
		return eval.NewThingEnv(state.GetNodeResult(id)), nil
	}
	evalCtx.Env["var"] = func(name string) (any, error) {
		return eval.NewThingEnv(state.GetTemporary(name)), nil
	}
	evalCtx.Patchers = append(evalCtx.Patchers, &nodeEvalPatcher{})

	return &FlowContext{
		Context: ctx,
		Cancel:  cancel,
		Data:    data,
		// Placeholders:      placeholders,
		EvalCtx:           evalCtx,
		FlowProviders:     providers,
		FlowContextLimits: limits,
		FlowContextState:  *state,
	}
}

type FlowContextData interface {
	CommandData() *discord.CommandInteraction
	MessageComponentData() discord.ComponentInteraction
	Interaction() *discord.InteractionEvent
	Event() ws.Event
	UserID() discord.UserID
	GuildID() discord.GuildID
	ChannelID() discord.ChannelID
}

type FlowContextLimits struct {
	MaxStackDepth int
	MaxOperations int
	MaxCredits    int

	stackDepth int
	operations int
	credits    int
}

func (c *FlowContextLimits) CreditsUsed() int {
	return c.credits
}

func (c *FlowContext) startOperation(credits int) error {
	if c.Err() != nil {
		return c.Err()
	}

	if err := c.increaseStackDepth(); err != nil {
		return err
	}

	if err := c.increaseOperations(); err != nil {
		return err
	}

	if err := c.increaseCredits(credits); err != nil {
		return err
	}

	return nil
}

func (c *FlowContext) endOperation() {
	c.decreaseStackDepth()
}

func (c *FlowContext) increaseStackDepth() error {
	c.stackDepth++
	if c.stackDepth > c.MaxStackDepth && c.MaxStackDepth != 0 {
		return &FlowError{
			Code:    FlowNodeErrorMaxStackDepthReached,
			Message: fmt.Sprintf("max stack depth reached: %d", c.MaxStackDepth),
		}
	}
	return nil
}

func (c *FlowContext) decreaseStackDepth() {
	if c.stackDepth > 0 {
		c.stackDepth--
	}
}

func (c *FlowContext) increaseOperations() error {
	c.operations++
	if c.operations > c.MaxOperations && c.MaxOperations != 0 {
		return &FlowError{
			Code:    FlowNodeErrorMaxOperationsReached,
			Message: fmt.Sprintf("operations limit exceeded: %d", c.MaxOperations),
		}
	}
	return nil
}

func (c *FlowContext) increaseCredits(credits int) error {
	c.credits += credits
	if c.credits > c.MaxCredits && c.MaxCredits != 0 {
		return &FlowError{
			Code:    FlowNodeErrorMaxCreditsReached,
			Message: fmt.Sprintf("credits limit exceeded: %d", c.MaxCredits),
		}
	}
	return nil
}

func (c *FlowContext) IsEntry() bool {
	return c.stackDepth == 1
}

func (c *FlowContext) suspend(t ResumePointType, resumePointID string, nodeID string) (*ResumePoint, error) {
	s, err := c.ResumePoint.CreateResumePoint(c.Context, ResumePoint{
		ID:     resumePointID,
		Type:   t,
		NodeID: nodeID,
		State:  c.FlowContextState.Copy(),
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}
