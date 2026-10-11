package webhook

import (
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

const EventType ws.EventType = "KITE_WEBHOOK"

const (
	// MaxBodySize bounds the body of a webhook request. The event is stored
	// with every resume point of the flow, so it has to stay small.
	MaxBodySize = 64 * 1024
	// MaxMetadataSize bounds the headers and query parameters together.
	MaxMetadataSize = 16 * 1024
)

// Event triggers the flow of a webhook event listener. It's a gateway event
// so it can go through the same execution path as Discord events.
type Event struct {
	// Headers of the request by lowercase name, without cookies.
	Headers map[string]string `json:"headers"`
	Query   map[string]string `json:"query"`
	Body    string            `json:"body"`
}

// Op is the dispatch op code, like every other event that triggers flows.
func (e *Event) Op() ws.OpCode { return 0 }

func (e *Event) EventType() ws.EventType { return EventType }
