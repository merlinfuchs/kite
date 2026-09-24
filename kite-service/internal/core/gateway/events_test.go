package gateway

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// Protocol frames report an empty event type. The dispatch path can never
// match one, so they are dropped at the gateway handler -- but only because
// nothing downstream uses an empty type as a real key. These tests pin that
// assumption, since it is what makes the filter safe.

func TestProtocolEventsReportEmptyEventType(t *testing.T) {
	// A representative sample of arikawa's protocol frames. If a future
	// version gives one of these a real event type the filter simply stops
	// dropping it, which is harmless. The dangerous direction is the reverse,
	// covered below.
	protocol := []ws.Event{
		&gateway.HeartbeatAckEvent{},
		&gateway.HelloEvent{},
		&gateway.ReconnectEvent{},
		func() *gateway.InvalidSessionEvent { e := gateway.InvalidSessionEvent(false); return &e }(),
	}

	for _, e := range protocol {
		if got := e.EventType(); got != "" {
			t.Errorf("%T reports event type %q, expected empty", e, got)
		}
	}
}

// The filter is only safe if no event the engine acts on carries an empty
// type. If one ever does, it would be silently swallowed at the gateway.
func TestDispatchedEventsHaveNonEmptyEventType(t *testing.T) {
	dispatched := []ws.Event{
		&gateway.MessageCreateEvent{},
		&gateway.MessageUpdateEvent{},
		&gateway.MessageDeleteEvent{},
		&gateway.GuildMemberAddEvent{},
		&gateway.GuildMemberRemoveEvent{},
		&gateway.MessageReactionAddEvent{},
		&gateway.InteractionCreateEvent{},
		&gateway.GuildCreateEvent{},
		&gateway.ReadyEvent{},
	}

	for _, e := range dispatched {
		if e.EventType() == "" {
			t.Errorf("%T reports an empty event type and would be filtered out", e)
		}
	}
}

// No event listener type may be empty, or the filter would drop events that
// listener was registered for.
func TestNoEventListenerTypeIsEmpty(t *testing.T) {
	types := []model.EventListenerType{
		model.EventListenerTypeDiscordMessageCreate,
		model.EventListenerTypeDiscordMessageUpdate,
		model.EventListenerTypeDiscordMessageDelete,
		model.EventListenerTypeDiscordGuildMemberAdd,
		model.EventListenerTypeDiscordGuildMemberRemove,
	}

	for _, tp := range types {
		if tp == "" {
			t.Error("an event listener type is empty, which the protocol filter would swallow")
		}
	}
}
