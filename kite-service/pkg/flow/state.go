package flow

import (
	"encoding/json"
	"slices"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type FlowContextState struct {
	NodeStates  map[string]*FlowContextNodeState `json:"node_states"`
	Temporaries map[string]thing.Thing           `json:"temporaries"`

	// Triggers holds the interactions or events of earlier executions, oldest
	// first. It's only set in resumed executions.
	Triggers []FlowTrigger `json:"triggers,omitempty"`

	// ResumeTrigger is the interaction or event a durable sleep continues
	// with. It's kept apart from Triggers, which are only earlier executions.
	ResumeTrigger *FlowTrigger `json:"resume_trigger,omitempty"`

	// DurableSleeps counts the durable sleeps of this execution, including the
	// ones it resumed from, so a cycle through a Wait block can't keep the flow
	// alive forever. Clicks and submits start a new execution from zero.
	DurableSleeps int `json:"durable_sleeps,omitempty"`
}

func NewFlowContextState() *FlowContextState {
	return &FlowContextState{
		NodeStates:  make(map[string]*FlowContextNodeState),
		Temporaries: make(map[string]thing.Thing),
	}
}

func (s FlowContextState) MarshalJSON() ([]byte, error) {
	// The alias drops this method so json.Marshal doesn't recurse.
	type state FlowContextState
	aux := state(s)
	aux.NodeStates = make(map[string]*FlowContextNodeState, len(s.NodeStates))
	aux.Temporaries = make(map[string]thing.Thing, len(s.Temporaries))

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
	// Triggers are never mutated, only replaced, so they can be shared.
	copy := *s
	copy.NodeStates = make(map[string]*FlowContextNodeState, len(s.NodeStates))
	copy.Temporaries = make(map[string]thing.Thing, len(s.Temporaries))

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

	// Copies share the slice, so it's rebuilt instead of appended to. The first
	// trigger is always kept since it started the flow.
	triggers := append(slices.Clone(s.Triggers), *trigger)
	if len(triggers) > maxStoredTriggers {
		triggers = slices.Delete(triggers, 1, len(triggers)-maxStoredTriggers+1)
	}
	s.Triggers = triggers
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
