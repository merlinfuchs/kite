package schedule

import (
	"time"

	"github.com/diamondburned/arikawa/v3/utils/ws"
)

const EventType ws.EventType = "KITE_SCHEDULE"

// Event triggers the flow of a scheduled event listener. It's a gateway event
// so it can go through the same execution path as Discord events.
type Event struct {
	// Time is the occurrence the run was scheduled for, not when it started.
	Time time.Time `json:"time"`
}

// Op is the dispatch op code, like every other event that triggers flows.
func (e *Event) Op() ws.OpCode { return 0 }

func (e *Event) EventType() ws.EventType { return EventType }
