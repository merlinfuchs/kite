package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

	EntryNodeID string

	Data    FlowContextData
	EvalCtx eval.Context
	Cancel  context.CancelFunc
}

func NewContext(
	ctx context.Context,
	timeout time.Duration,
	entryNodeID string,
	data FlowContextData,
	providers FlowProviders,
	limits FlowContextLimits,
	evalCtx eval.Context,
	state *FlowContextState,
) *FlowContext {
	ctx, cancel := context.WithTimeout(ctx, timeout)

	if state == nil {
		state = &FlowContextState{
			NodeStates: make(map[string]*FlowContextNodeState),
		}
	}

	nodeEvalEnv := &nodeEvalEnv{
		state: state,
	}

	evalCtx.Env["node"] = nodeEvalEnv.GetNode
	evalCtx.Env["result"] = func(id string) (any, error) {
		node, err := nodeEvalEnv.GetNode(id)
		if err != nil {
			return nil, err
		}
		if node == nil {
			return nil, nil
		}
		return node.(map[string]any)["result"], nil
	}
	evalCtx.Patchers = append(evalCtx.Patchers, &nodeEvalPatcher{})

	return &FlowContext{
		Context:     ctx,
		Cancel:      cancel,
		EntryNodeID: entryNodeID,
		Data:        data,
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

func (c *FlowContext) SetEntryNodeID(nodeID string) {
	c.EntryNodeID = nodeID
}

func (c *FlowContext) suspend(t ResumePointType, nodeID string) (*ResumePoint, error) {
	s, err := c.ResumePoint.CreateResumePoint(c.Context, ResumePoint{
		Type:   t,
		NodeID: nodeID,
		State:  c.FlowContextState.Copy(),
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

type FlowContextState struct {
	NodeStates map[string]*FlowContextNodeState `json:"node_states"`
}

func (s FlowContextState) MarshalJSON() ([]byte, error) {
	aux := struct {
		NodeStates map[string]*FlowContextNodeState `json:"node_states"`
	}{
		NodeStates: make(map[string]*FlowContextNodeState, len(s.NodeStates)),
	}
	// We don't want to serialize empty node states
	for k, v := range s.NodeStates {
		if !v.IsEmpty() {
			aux.NodeStates[k] = v
		}
	}

	return json.Marshal(aux)
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

func (s *FlowContextState) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

func (s *FlowContextState) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

type FlowContextNodeState struct {
	ConditionBaseValue thing.Any `json:"condition_base_value,omitempty"`
	ConditionItemMet   bool      `json:"condition_item_met,omitempty"`
	Result             thing.Any `json:"result,omitempty"`
	LoopExited         bool      `json:"loop_exited,omitempty"`
}

func (s *FlowContextNodeState) MarshalJSON() ([]byte, error) {
	// We want to ommit nil values
	aux := struct {
		ConditionBaseValue any  `json:"condition_base_value,omitempty"`
		ConditionItemMet   bool `json:"condition_item_met,omitempty"`
		Result             any  `json:"result,omitempty"`
		LoopExited         bool `json:"loop_exited,omitempty"`
	}{
		ConditionBaseValue: s.ConditionBaseValue.Inner,
		ConditionItemMet:   s.ConditionItemMet,
		Result:             s.Result.Inner,
		LoopExited:         s.LoopExited,
	}

	return json.Marshal(aux)
}

func (s *FlowContextNodeState) IsEmpty() bool {
	return s.ConditionBaseValue.IsNil() &&
		!s.ConditionItemMet &&
		s.Result.IsNil() &&
		!s.LoopExited
}

func (s *FlowContextNodeState) Copy() *FlowContextNodeState {
	return &FlowContextNodeState{
		ConditionBaseValue: s.ConditionBaseValue,
		ConditionItemMet:   s.ConditionItemMet,
		Result:             s.Result,
		LoopExited:         s.LoopExited,
	}
}
