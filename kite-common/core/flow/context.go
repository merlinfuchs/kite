package flow

import (
	"context"
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type FlowContext struct {
	context.Context
	FlowProviders
	FlowContextLimits

	Data      FlowContextData
	Tempories FlowContextTemporaries
	Variables FlowContextVariables
}

func NewContext(ctx context.Context, data FlowContextData, providers FlowProviders, limits FlowContextLimits) *FlowContext {
	return &FlowContext{
		Context: ctx,
		Data:    data,
		Variables: FlowContextVariables{
			Variables: make(map[string]FlowValue),
		},
		FlowProviders:     providers,
		FlowContextLimits: limits,
	}
}

type FlowContextData interface {
	CommandData() *discord.CommandInteraction
	MessageComponentData() discord.ComponentInteraction
	Interaction() *discord.InteractionEvent
	EventData() gateway.Event
	GuildID() string
	ChannelID() string
}

type FlowContextVariables struct {
	Variables map[string]FlowValue
}

func (v *FlowContextVariables) SetVariable(name string, value FlowValue) {
	v.Variables[name] = value
}

func (v *FlowContextVariables) Variable(name string) FlowValue {
	value := v.Variables[name]
	return value
}

type FlowContextTemporaries struct {
	ConditionBaseValue     FlowValue
	ConditionItemMet       bool
	ConditionAllowMultiple bool
}

func (t *FlowContextTemporaries) InitCondition(baseValue FlowValue, allowMultiple bool) {
	t.ConditionBaseValue = baseValue
	t.ConditionAllowMultiple = allowMultiple
	t.ConditionItemMet = false
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
