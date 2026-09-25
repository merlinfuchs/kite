package flow

import (
	"encoding/json"
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
)

// maxStoredTriggers bounds how many triggers a resume point keeps: the one that
// started the flow and the most recent ones. A modal submission can hold up to
// 20 KB of text.
const maxStoredTriggers = 4

// FlowTrigger is the interaction or event that started an execution. Resume
// points store it so sub-flows can still reach it after the flow resumes with
// a different interaction.
type FlowTrigger struct {
	Interaction *discord.InteractionEvent
	Event       ws.Event
}

func newFlowTrigger(data FlowContextData) *FlowTrigger {
	if i := data.Interaction(); i != nil {
		stored := *i
		// The token is a credential and the message and channel aren't
		// reachable from placeholders, so none of them are worth storing.
		stored.Token = ""
		stored.Message = nil
		stored.Channel = nil
		return &FlowTrigger{Interaction: &stored}
	}

	if e := data.Event(); e != nil {
		return &FlowTrigger{Event: e}
	}

	return nil
}

type flowTriggerJSON struct {
	Interaction *discord.InteractionEvent `json:"interaction,omitempty"`
	EventOp     ws.OpCode                 `json:"event_op,omitempty"`
	EventType   ws.EventType              `json:"event_type,omitempty"`
	Event       json.RawMessage           `json:"event,omitempty"`
}

func (t FlowTrigger) MarshalJSON() ([]byte, error) {
	aux := flowTriggerJSON{Interaction: t.Interaction}

	if t.Event != nil {
		data, err := json.Marshal(t.Event)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal event: %w", err)
		}

		aux.EventOp = t.Event.Op()
		aux.EventType = t.Event.EventType()
		aux.Event = data
	}

	return json.Marshal(aux)
}

func (t *FlowTrigger) UnmarshalJSON(data []byte) error {
	var aux flowTriggerJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	t.Interaction = aux.Interaction

	if aux.EventType != "" {
		var event ws.Event
		if aux.EventType == schedule.EventType {
			// Kite's own event, which arikawa doesn't know about.
			event = new(schedule.Event)
		} else {
			newEvent := gateway.OpUnmarshalers.Lookup(aux.EventOp, aux.EventType)
			if newEvent == nil {
				return fmt.Errorf("unknown event type: %s", aux.EventType)
			}
			event = newEvent()
		}

		if err := json.Unmarshal(aux.Event, event); err != nil {
			return fmt.Errorf("failed to unmarshal event: %w", err)
		}
		t.Event = event
	}

	return nil
}
