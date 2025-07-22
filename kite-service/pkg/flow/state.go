package flow

import (
	"encoding/json"
	"fmt"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type FlowContextState struct {
	NodeStates map[string]*FlowContextNodeState `json:"node_states"`
	ResultKeys map[string]string                `json:"result_keys"`
}

func NewFlowContextState() *FlowContextState {
	return &FlowContextState{
		NodeStates: make(map[string]*FlowContextNodeState),
		ResultKeys: make(map[string]string),
	}
}

func (s FlowContextState) MarshalJSON() ([]byte, error) {
	aux := struct {
		NodeStates map[string]*FlowContextNodeState `json:"node_states"`
		ResultKeys map[string]string                `json:"result_keys"`
	}{
		NodeStates: make(map[string]*FlowContextNodeState, len(s.NodeStates)),
		ResultKeys: s.ResultKeys,
	}
	// We don't want to serialize empty node states
	for k, v := range s.NodeStates {
		if !v.IsEmpty() {
			aux.NodeStates[k] = v
		}
	}

	return json.Marshal(aux)
}

func (s *FlowContextState) GetNodeState(node *CompiledFlowNode) *FlowContextNodeState {
	// Remember the result key of the node so we can access it later
	if node.Data.ResultKey != "" {
		s.ResultKeys[node.Data.ResultKey] = node.ID
	}

	state, ok := s.NodeStates[node.ID]
	if !ok {
		state = &FlowContextNodeState{}
		s.NodeStates[node.ID] = state
	}

	return state
}

func (s *FlowContextState) GetNodeResultByID(id string) thing.Thing {
	state := s.NodeStates[id]
	if state == nil {
		return thing.Null
	}

	return state.Result
}

func (s *FlowContextState) GetNodeResultByKey(key string) thing.Thing {
	fmt.Println("GetNodeResultByKey", key)
	fmt.Println("ResultKeys", s.ResultKeys)
	if id, ok := s.ResultKeys[key]; ok {
		return s.GetNodeResultByID(id)
	}

	return thing.Null
}

func (s *FlowContextState) Copy() FlowContextState {
	copy := FlowContextState{
		NodeStates: make(map[string]*FlowContextNodeState, len(s.NodeStates)),
		ResultKeys: make(map[string]string, len(s.ResultKeys)),
	}

	for k, v := range s.NodeStates {
		if !v.IsEmpty() {
			copy.NodeStates[k] = v.Copy()
		}
	}

	for k, v := range s.ResultKeys {
		copy.ResultKeys[k] = v
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
