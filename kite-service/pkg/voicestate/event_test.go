package voicestate

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

func TestEventAction(t *testing.T) {
	tests := []struct {
		name string
		old  discord.ChannelID
		new  discord.ChannelID
		want Action
	}{
		{"join", 0, 1, ActionJoin},
		{"leave", 1, 0, ActionLeave},
		{"leave with null channel", 1, discord.NullChannelID, ActionLeave},
		{"leave without cached state", 0, 0, ActionLeave},
		{"move", 1, 2, ActionMove},
		{"mute in the same channel", 1, 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Event{
				VoiceState:   discord.VoiceState{ChannelID: tt.new},
				OldChannelID: tt.old,
			}
			if got := e.Action(); got != tt.want {
				t.Errorf("Action() = %q, want %q", got, tt.want)
			}
		})
	}
}
