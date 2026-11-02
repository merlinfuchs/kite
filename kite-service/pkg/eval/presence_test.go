package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

func TestPresencePlaceholders(t *testing.T) {
	custom := func(text string, emoji *discord.Emoji) discord.Activity {
		return discord.Activity{Name: "Custom Status", Type: discord.CustomActivity, State: text, Emoji: emoji}
	}

	tests := map[string]struct {
		presence discord.Presence
		previous *discord.Presence
		want     string
	}{
		"custom status set": {
			discord.Presence{Status: discord.OnlineStatus, Activities: []discord.Activity{
				{Name: "Minecraft", Type: discord.GameActivity},
				custom("discord.gg/kite", &discord.Emoji{Name: "🪁"}),
			}},
			&discord.Presence{Status: discord.OnlineStatus},
			"online|online|discord.gg/kite|🪁|",
		},
		"custom emoji": {
			discord.Presence{Status: discord.IdleStatus, Activities: []discord.Activity{
				custom("", &discord.Emoji{ID: 5, Name: "kite"}),
			}},
			&discord.Presence{Status: discord.OnlineStatus, Activities: []discord.Activity{custom("hi", nil)}},
			"idle|online||<:kite:5>|hi",
		},
		// Members who aren't cached were offline.
		"comes online": {
			discord.Presence{Status: discord.DoNotDisturbStatus},
			nil,
			"dnd|offline|||",
		},
		"goes offline": {
			discord.Presence{Status: discord.OfflineStatus},
			&discord.Presence{Status: discord.InvisibleStatus},
			"offline|offline|||",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.presence.User = discord.User{ID: 7}
			tt.presence.GuildID = testGuildID
			env := NewEventEnv(&guildevent.PresenceChange{
				PresenceUpdateEvent: gateway.PresenceUpdateEvent{Presence: tt.presence},
				Previous:            tt.previous,
			}, nil)
			ctx := NewContext(Env{"user": env.User, "presence": env.Presence})

			got, err := EvalTemplateToString(t.Context(),
				"{{presence}}|{{presence.previous_status}}|{{presence.custom_status}}|{{presence.custom_status_emoji}}|{{presence.previous_custom_status}}",
				ctx,
			)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}

			id, err := EvalTemplateToString(t.Context(), "{{user.id}}", ctx)
			if err != nil || id != "7" {
				t.Errorf("user.id = %q, %v, want 7", id, err)
			}
		})
	}
}
