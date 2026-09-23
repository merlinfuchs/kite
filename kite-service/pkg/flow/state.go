package flow

import (
	"encoding/json"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type FlowContextState struct {
	NodeStates  map[string]*FlowContextNodeState `json:"node_states"`
	Temporaries map[string]thing.Thing           `json:"temporaries"`

	// Origin and Previous are only set in resumed executions. Origin started
	// the flow, Previous started the execution that created the resume point.
	Origin   *FlowTrigger `json:"origin,omitempty"`
	Previous *FlowTrigger `json:"previous,omitempty"`
	// ModalInputs holds the inputs of earlier modal submissions, newest first.
	ModalInputs []map[string]string `json:"modal_inputs,omitempty"`
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
		Origin      *FlowTrigger                     `json:"origin,omitempty"`
		Previous    *FlowTrigger                     `json:"previous,omitempty"`
		ModalInputs []map[string]string              `json:"modal_inputs,omitempty"`
	}{
		NodeStates:  make(map[string]*FlowContextNodeState, len(s.NodeStates)),
		Temporaries: make(map[string]thing.Thing, len(s.Temporaries)),
		Origin:      s.Origin,
		Previous:    s.Previous,
		ModalInputs: s.ModalInputs,
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
		// Triggers and inputs are never mutated, only replaced.
		Origin:      s.Origin,
		Previous:    s.Previous,
		ModalInputs: s.ModalInputs,
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

// recordTrigger stores the trigger of the current execution before the state is
// saved in a resume point.
func (s *FlowContextState) recordTrigger(data FlowContextData) {
	trigger := newFlowTrigger(data)
	if trigger == nil {
		return
	}

	if s.Origin == nil {
		s.Origin = trigger
	}
	s.Previous = trigger

	if inputs := trigger.modalInputs(); inputs != nil {
		modalInputs := append([]map[string]string{inputs}, s.ModalInputs...)
		s.ModalInputs = modalInputs[:min(len(modalInputs), maxStoredModalInputs)]
	}
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
