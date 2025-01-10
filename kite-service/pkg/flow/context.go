package flow

import (
	"context"
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type FlowContext struct {
	context.Context
	FlowProviders
	FlowContextLimits
	FlowContextState

	Data    FlowContextData
	EvalCtx eval.Context
}

func NewContext(
	ctx context.Context,
	data FlowContextData,
	providers FlowProviders,
	limits FlowContextLimits,
	evalCtx eval.Context,
) *FlowContext {
	state := FlowContextState{
		NodeStates: make(map[string]*FlowContextNodeState),
	}

	evalCtx.Env["node"] = (&nodeEvalEnv{
		state: &state,
	}).GetNode
	evalCtx.Patchers = append(evalCtx.Patchers, &nodeEvalPatcher{})

	return &FlowContext{
		Context: ctx,
		Data:    data,
		// Placeholders:      placeholders,
		EvalCtx:           evalCtx,
		FlowProviders:     providers,
		FlowContextLimits: limits,
		FlowContextState:  state,
	}
}

type FlowContextData interface {
	CommandData() *discord.CommandInteraction
	MessageComponentData() discord.ComponentInteraction
	Interaction() *discord.InteractionEvent
	Event() ws.Event
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

type FlowContextState struct {
	NodeStates map[string]*FlowContextNodeState
}

func (c *FlowContextState) GetNodeState(nodeID string) *FlowContextNodeState {
	state, ok := c.NodeStates[nodeID]
	if !ok {
		state = &FlowContextNodeState{}
		c.NodeStates[nodeID] = state
	}

	return state
}

type FlowContextNodeState struct {
	ConditionBaseValue thing.Any
	ConditionItemMet   bool
	Result             thing.Any
	LoopExited         bool
}
