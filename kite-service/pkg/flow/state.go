package flow

import (
	"encoding/json"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

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
		if !v.IsEmpty() {
			copy.NodeStates[k] = v.Copy()
		}
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
	ConditionBaseValue thing.Thing `json:"condition_base_value,omitzero"`
	ConditionItemMet   bool        `json:"condition_item_met,omitempty"`
	Result             thing.Thing `json:"result,omitzero"`
	LoopExited         bool        `json:"loop_exited,omitempty"`
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
