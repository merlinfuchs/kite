package engine

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/voicestate"
)

// Every guild the bot is in arrives as a raw GUILD_CREATE on connect, so only
// arikawa's derived join and leave events may trigger the bot server listeners.
func TestShouldHandleGuildJoinAndLeave(t *testing.T) {
	tests := []struct {
		name  string
		event ws.Event
		want  bool
	}{
		{"raw guild create", &gateway.GuildCreateEvent{}, false},
		{"raw guild delete", &gateway.GuildDeleteEvent{}, false},
		{"guild ready on connect", &state.GuildReadyEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}}, false},
		{"guild available after outage", &state.GuildAvailableEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}}, false},
		{"guild unavailable", &state.GuildUnavailableEvent{GuildDeleteEvent: &gateway.GuildDeleteEvent{Unavailable: true}}, false},
		{"guild join", &state.GuildJoinEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}}, true},
		{"guild leave", &state.GuildLeaveEvent{GuildDeleteEvent: &gateway.GuildDeleteEvent{}}, true},
	}

	l := &EventListener{}
	for _, tt := range tests {
		if got := l.shouldHandleEvent(tt.event); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

// Muting, deafening and streaming update the voice state too. Only changing
// the channel may trigger the voice state listeners.
func TestShouldHandleVoiceStateUpdate(t *testing.T) {
	voice := func(oldChannelID, channelID discord.ChannelID, bot bool) *voicestate.Event {
		return &voicestate.Event{
			VoiceState: discord.VoiceState{
				ChannelID: channelID,
				Member:    &discord.Member{User: discord.User{Bot: bot}},
			},
			OldChannelID: oldChannelID,
		}
	}

	tests := []struct {
		name  string
		event ws.Event
		want  bool
	}{
		{"raw voice state update", &gateway.VoiceStateUpdateEvent{}, false},
		{"join", voice(0, 1, false), true},
		{"leave", voice(1, 0, false), true},
		{"move", voice(1, 2, false), true},
		{"mute in the same channel", voice(1, 1, false), false},
		{"bot join", voice(0, 1, true), false},
		{"join without member", &voicestate.Event{VoiceState: discord.VoiceState{ChannelID: 1}}, true},
	}

	l := &EventListener{}
	for _, tt := range tests {
		if got := l.shouldHandleEvent(tt.event); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}
