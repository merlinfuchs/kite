package flow

import (
	"encoding/json"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type FlowContextState struct {
	NodeStates  map[string]*FlowContextNodeState `json:"node_states"`
	Temporaries map[string]thing.Thing           `json:"temporaries"`
}

func NewFlowContextState() *FlowContextState {
	return &FlowContextState{
		NodeStates:  make(map[string]*FlowContextNodeState),
		Temporaries: make(map[string]thing.Thing),
	}
}

func (s FlowContextState) MarshalJSON() ([]byte, error) {
	aux := struct {
		NodeStates  map[string]*FlowContextNodeState `json:"node_states"`
		Temporaries map[string]thing.Thing           `json:"temporaries"`
	}{
		NodeStates:  make(map[string]*FlowContextNodeState, len(s.NodeStates)),
		Temporaries: make(map[string]thing.Thing, len(s.Temporaries)),
	}
	// We don't want to serialize empty node states
	for k, v := range s.NodeStates {
		if !v.IsEmpty() {
			aux.NodeStates[k] = v
		}
	}

	for k, v := range s.Temporaries {
		if !v.IsNil() {
			aux.Temporaries[k] = v
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

func (s *FlowContextState) GetNodeResult(id string) thing.Thing {
	state := s.NodeStates[id]
	if state == nil {
		return thing.Null
	}

	return state.Result
}

func (s *FlowContextState) StoreNodeResult(node *CompiledFlowNode, result thing.Thing) {
	state := s.GetNodeState(node.ID)
	state.Result = result
	if node.Data.TemporaryName != "" {
		s.Temporaries[node.Data.TemporaryName] = result
	}
}

func (s *FlowContextState) StoreNodeBaseValue(node *CompiledFlowNode, value thing.Thing) {
	state := s.GetNodeState(node.ID)
	state.ConditionBaseValue = value
}

func (s *FlowContextState) GetTemporary(name string) thing.Thing {
	if v, ok := s.Temporaries[name]; ok {
		return v
	}

	return thing.Null
}

func (s *FlowContextState) SetTemporary(name string, value thing.Thing) {
	s.Temporaries[name] = value
}

func (s *FlowContextState) Copy() FlowContextState {
	copy := FlowContextState{
		NodeStates:  make(map[string]*FlowContextNodeState, len(s.NodeStates)),
		Temporaries: make(map[string]thing.Thing, len(s.Temporaries)),
	}

	for k, v := range s.NodeStates {
		if !v.IsEmpty() {
			copy.NodeStates[k] = v.Copy()
		}
	}

	for k, v := range s.Temporaries {
		if !v.IsNil() {
			copy.Temporaries[k] = v
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
