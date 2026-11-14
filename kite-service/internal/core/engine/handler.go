package engine

import (
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/core/command"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// Commands and Discord events are routed to handlers by what each of them
// declares. The flows of commands and event listeners are handlers, and
// modules will be too, so routing doesn't depend on what defines them.
// Schedules, webhooks and resume points still go to flows directly, and
// plugins have their own path.

// commandHandler handles the commands it declares.
type commandHandler interface {
	commandKeys() []command.Key
	handleCommand(session *state.State, event *gateway.InteractionCreateEvent)
}

// eventHandler handles Discord events of the types it declares, once
// shouldHandleEvent let them through for the type.
type eventHandler interface {
	eventTypes() []model.EventListenerType
	handleEvent(eventType model.EventListenerType, session *state.State, event gateway.Event)
}

// eventRoute is a handler of an event type, which it gets events as. A boost
// message goes to the routes of message create and of boosts.
type eventRoute struct {
	eventType model.EventListenerType
	handler   eventHandler
}
