package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type FlowContext struct {
	context.Context
	waitGroup *sync.WaitGroup

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
		Context:   ctx,
		waitGroup: &sync.WaitGroup{},

		Data: data,
		// Placeholders:      placeholders,
		EvalCtx:           evalCtx,
		FlowProviders:     providers,
		FlowContextLimits: limits,
		FlowContextState:  state,
	}
}

func (c *FlowContext) Copy() *FlowContext {
	return &FlowContext{
		Context:           c.Context,
		waitGroup:         c.waitGroup,
		Data:              c.Data,
		EvalCtx:           c.EvalCtx,
		FlowProviders:     c.FlowProviders,
		FlowContextLimits: c.FlowContextLimits,
		FlowContextState:  c.FlowContextState.Copy(),
	}
}

func (c *FlowContext) Wait() {
	c.waitGroup.Wait()
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

func (s *FlowContextState) GetNodeState(nodeID string) *FlowContextNodeState {
	state, ok := s.NodeStates[nodeID]
	if !ok {
		state = &FlowContextNodeState{}
		s.NodeStates[nodeID] = state
	}

	return state
}

func (s *FlowContextState) Copy() FlowContextState {
	copy := FlowContextState{
		NodeStates: make(map[string]*FlowContextNodeState, len(s.NodeStates)),
	}

	for k, v := range s.NodeStates {
		copy.NodeStates[k] = v.Copy()
	}

	return copy
}

func (s *FlowContextNodeState) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

func (s *FlowContextNodeState) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

type FlowContextNodeState struct {
	ConditionBaseValue thing.Any `json:"condition_base_value"`
	ConditionItemMet   bool      `json:"condition_item_met"`
	Result             thing.Any `json:"result"`
	LoopExited         bool      `json:"loop_exited"`
}

func (s *FlowContextNodeState) Copy() *FlowContextNodeState {
	return &FlowContextNodeState{
		ConditionBaseValue: s.ConditionBaseValue,
		ConditionItemMet:   s.ConditionItemMet,
		Result:             s.Result,
		LoopExited:         s.LoopExited,
	}
}
