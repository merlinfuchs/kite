package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

func TestVoicePlaceholders(t *testing.T) {
	session := testSession(t)
	const template = "{{voice.action}}|{{channel.id}}|{{voice.old_channel.id}}"
	tests := map[string]struct {
		channelID, oldChannelID discord.ChannelID
		action                  string
		want                    string
	}{
		"join":  {21, 0, "join", "join|21|"},
		"move":  {22, 21, "move", "move|22|21"},
		"leave": {0, 21, "leave", "leave|21|21"},
	}
	for name, tt := range tests {
		env := NewEventEnv(&guildevent.VoiceStateUpdate{
			VoiceStateUpdateEvent: gateway.VoiceStateUpdateEvent{VoiceState: discord.VoiceState{
				GuildID: testGuildID, ChannelID: tt.channelID, UserID: testUserID,
			}},
			OldChannelID: tt.oldChannelID,
			Action:       tt.action,
		}, session)
		c := NewContext(Env{"channel": env.Channel, "voice": env.Voice})
		if got := evalString(t, c, template); got != tt.want {
			t.Errorf("%s: got %q, want %q", name, got, tt.want)
		}
	}
}
