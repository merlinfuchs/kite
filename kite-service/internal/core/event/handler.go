package event

import (
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type EventHandler interface {
	HandleEvent(appID string, session *state.State, event gateway.Event)
}

type EventHandlerWrapper struct {
	next EventHandler

	messageInstanceStore store.MessageInstanceStore
}

func NewEventHandlerWrapper(
	next EventHandler,
	messageInstanceStore store.MessageInstanceStore,
) *EventHandlerWrapper {
	return &EventHandlerWrapper{
		next:                 next,
		messageInstanceStore: messageInstanceStore,
	}
}

func (h *EventHandlerWrapper) HandleEvent(appID string, session *state.State, event gateway.Event) {
	switch e := event.(type) {
	case *gateway.MessageDeleteEvent:
		h.HandleMessageDeleteEvent(appID, session, e)
	}

	h.next.HandleEvent(appID, session, event)
}
