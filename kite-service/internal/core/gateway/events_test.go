package gateway

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
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
		&gateway.VoiceStateUpdateEvent{},
		&gateway.MessageReactionAddEvent{},
		&gateway.MessageReactionRemoveEvent{},
		&gateway.InteractionCreateEvent{},
		&gateway.GuildCreateEvent{},
		&gateway.GuildDeleteEvent{},
		&state.GuildJoinEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}},
		&state.GuildLeaveEvent{GuildDeleteEvent: &gateway.GuildDeleteEvent{}},
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
		model.EventListenerTypeDiscordMessageReactionAdd,
		model.EventListenerTypeDiscordMessageReactionRemove,
		model.EventListenerTypeDiscordGuildCreate,
		model.EventListenerTypeDiscordGuildDelete,
		model.EventListenerTypeDiscordVoiceChannelJoin,
		model.EventListenerTypeDiscordVoiceChannelLeave,
	}

	for _, tp := range types {
		if tp == "" {
			t.Error("an event listener type is empty, which the protocol filter would swallow")
		}
	}
}

func TestVoiceStateUpdateEventKeepsItsDiscordEvent(t *testing.T) {
	original := &gateway.VoiceStateUpdateEvent{}
	event := &voiceStateUpdateEvent{
		VoiceStateUpdateEvent: original,
		eventType:             ws.EventType(model.EventListenerTypeDiscordVoiceChannelJoin),
	}

	if got := event.EventType(); got != ws.EventType(model.EventListenerTypeDiscordVoiceChannelJoin) {
		t.Errorf("EventType() = %q, want voice channel join", got)
	}
	if got := event.OriginalEvent(); got != original {
		t.Error("OriginalEvent() did not return the Discord voice state event")
	}
}

func TestVoiceStateTransitionType(t *testing.T) {
	tests := []struct {
		name     string
		previous discord.ChannelID
		current  discord.ChannelID
		want     model.EventListenerType
	}{
		{
			name:    "join",
			current: 10,
			want:    model.EventListenerTypeDiscordVoiceChannelJoin,
		},
		{
			name:     "leave",
			previous: 10,
			want:     model.EventListenerTypeDiscordVoiceChannelLeave,
		},
		{name: "same channel update", previous: 10, current: 10},
		{name: "move between channels", previous: 10, current: 11},
		{name: "remains disconnected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := voiceStateTransitionType(tt.previous, tt.current); got != tt.want {
				t.Errorf("voiceStateTransitionType() = %q, want %q", got, tt.want)
			}
		})
	}
}
