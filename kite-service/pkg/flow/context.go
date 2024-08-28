package flow

import (
	"context"
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/placeholder"
)

type FlowContext struct {
	context.Context
	FlowProviders
	FlowContextLimits
	FlowContextState

	Data         FlowContextData
	Placeholders *placeholder.Engine
}

func NewContext(
	ctx context.Context,
	data FlowContextData,
	providers FlowProviders,
	limits FlowContextLimits,
	placeholders *placeholder.Engine,
) *FlowContext {
	if data.Interaction() != nil {
		placeholders.AddProvider("interaction", placeholder.NewInteractionProvider(data.Interaction()))
	}

	placeholders.AddProvider("variables", &variablePlaceholderProvider{
		Variable: providers.Variable,
	})

	state := FlowContextState{
		NodeStates: make(map[string]*FlowContextNodeState),
	}

	flowStatePlaceHolderProvider := &flowStatePlaceholderProvider{
		state: &state,
	}
	placeholders.AddProvider("nodes", flowStatePlaceHolderProvider)

	return &FlowContext{
		Context:           ctx,
		Data:              data,
		Placeholders:      placeholders,
		FlowProviders:     providers,
		FlowContextLimits: limits,
		FlowContextState:  state,
	}
}

type FlowContextData interface {
	CommandData() *discord.CommandInteraction
	MessageComponentData() discord.ComponentInteraction
	Interaction() *discord.InteractionEvent
	EventData() gateway.Event
	GuildID() discord.GuildID
	ChannelID() discord.ChannelID
}

type FlowContextLimits struct {
	MaxStackDepth int
	MaxOperations int
	MaxActions    int

	stackDepth int
	operations int
	actions    int
}

func (c *FlowContext) startOperation() error {
	if c.Err() != nil {
		return c.Err()
	}

	if err := c.increaseStackDepth(); err != nil {
		return err
	}

	if err := c.increaseOperations(); err != nil {
		return err
	}

	return nil
}

func (c *FlowContext) endOperation() {
	c.decreaseStackDepth()
}

func (c *FlowContext) increaseStackDepth() error {
	c.stackDepth++
	if c.stackDepth > c.MaxStackDepth {
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
	if c.operations > c.MaxOperations {
		return &FlowError{
			Code:    FlowNodeErrorMaxOperationsReached,
			Message: fmt.Sprintf("operations limit exceeded: %d", c.MaxOperations),
		}
	}
	return nil
}

func (c *FlowContext) startAction() error {
	c.actions++
	if c.actions > c.MaxActions {
		return &FlowError{
			Code:    FlowNodeErrorMaxActionsReached,
			Message: fmt.Sprintf("actions limit exceeded: %d", c.MaxActions),
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
	ConditionBaseValue FlowString
	ConditionItemMet   bool
	Result             FlowValue
	LoopExited         bool
}
