package gateway

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// allPrivilegedFlags is what an app looks like after following the getting
// started guide, which tells every user to enable all three privileged
// intents in the developer portal. Most apps in production look like this,
// which is why intents have to be driven by what the app uses rather than by
// what it is permitted to have.
const allPrivilegedFlags = discord.ApplicationFlags(
	GATEWAY_MESSAGE_CONTENT | GATEWAY_GUILD_MEMBERS,
)

func TestIntentsForRequirements(t *testing.T) {
	tests := []struct {
		name  string
		reqs  model.AppGatewayRequirements
		flags discord.ApplicationFlags
		want  gateway.Intents
	}{
		{
			// The large majority of apps. Interactions are delivered without
			// any intent, so a command-only app needs nothing beyond the
			// guild intent that backs the dashboard's channel picker.
			name:  "command-only app gets guilds only despite all flags",
			reqs:  model.AppGatewayRequirements{},
			flags: allPrivilegedFlags,
			want:  gateway.IntentGuilds,
		},
		{
			name: "message listener adds guild messages and content",
			reqs: model.AppGatewayRequirements{
				EventListenerTypes: []model.EventListenerType{
					model.EventListenerTypeDiscordMessageCreate,
				},
			},
			flags: allPrivilegedFlags,
			want:  gateway.IntentGuilds | gateway.IntentGuildMessages | gateway.IntentMessageContent,
		},
		{
			// Requesting a privileged intent the app was never approved for
			// is rejected by Discord with close code 4014.
			name: "message content withheld without the portal flag",
			reqs: model.AppGatewayRequirements{
				EventListenerTypes: []model.EventListenerType{
					model.EventListenerTypeDiscordMessageCreate,
				},
			},
			flags: 0,
			want:  gateway.IntentGuilds | gateway.IntentGuildMessages,
		},
		{
			name: "member listener adds guild members",
			reqs: model.AppGatewayRequirements{
				EventListenerTypes: []model.EventListenerType{
					model.EventListenerTypeDiscordGuildMemberAdd,
				},
			},
			flags: allPrivilegedFlags,
			want:  gateway.IntentGuilds | gateway.IntentGuildMembers,
		},
		{
			name: "guild members withheld without the portal flag",
			reqs: model.AppGatewayRequirements{
				EventListenerTypes: []model.EventListenerType{
					model.EventListenerTypeDiscordGuildMemberAdd,
				},
			},
			flags: 0,
			want:  gateway.IntentGuilds,
		},
		{
			// Reactions were previously requested unconditionally for every
			// app even though no event listener type covers them.
			name: "starboard plugin adds reactions but not messages",
			reqs: model.AppGatewayRequirements{
				PluginEventTypes: []ws.EventType{"MESSAGE_REACTION_ADD"},
			},
			flags: allPrivilegedFlags,
			want:  gateway.IntentGuilds | gateway.IntentGuildMessageReactions,
		},
		{
			name: "counting plugin adds guild messages",
			reqs: model.AppGatewayRequirements{
				PluginEventTypes: []ws.EventType{"MESSAGE_CREATE"},
			},
			flags: allPrivilegedFlags,
			want:  gateway.IntentGuilds | gateway.IntentGuildMessages | gateway.IntentMessageContent,
		},
		{
			// Transitional: MESSAGE_DELETE drives message_instances cleanup,
			// and that needs the guild messages intent.
			name: "message instances keep guild messages on",
			reqs: model.AppGatewayRequirements{
				HasMessageInstances: true,
			},
			flags: allPrivilegedFlags,
			want:  gateway.IntentGuilds | gateway.IntentGuildMessages | gateway.IntentMessageContent,
		},
		{
			name: "requirements combine",
			reqs: model.AppGatewayRequirements{
				EventListenerTypes: []model.EventListenerType{
					model.EventListenerTypeDiscordMessageCreate,
					model.EventListenerTypeDiscordGuildMemberRemove,
				},
				PluginEventTypes: []ws.EventType{"MESSAGE_REACTION_ADD"},
			},
			flags: allPrivilegedFlags,
			want: gateway.IntentGuilds | gateway.IntentGuildMessages | gateway.IntentMessageContent |
				gateway.IntentGuildMembers | gateway.IntentGuildMessageReactions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intentsForRequirements(tt.reqs, tt.flags); got != tt.want {
				t.Errorf("intentsForRequirements() = %d, want %d (diff %d)",
					got, tt.want, got^tt.want)
			}
		})
	}
}

// Every app keeps the guild intent, because the dashboard's guild and channel
// pickers read from the state cache it populates.
func TestIntentsAlwaysIncludeGuilds(t *testing.T) {
	reqs := []model.AppGatewayRequirements{
		{},
		{HasMessageInstances: true},
		{EventListenerTypes: []model.EventListenerType{model.EventListenerTypeDiscordMessageDelete}},
		{PluginEventTypes: []ws.EventType{"MESSAGE_REACTION_ADD"}},
	}

	for i, r := range reqs {
		got := intentsForRequirements(r, allPrivilegedFlags)
		if got&gateway.IntentGuilds == 0 {
			t.Errorf("requirements[%d]: guild intent missing from %d", i, got)
		}
	}
}

// The fallback used when requirements cannot be loaded must stay at least as
// broad as the old unconditional behaviour, so a database blip degrades to
// wasteful rather than to dropping events.
func TestAllPermittedIntentsIsBroad(t *testing.T) {
	got := allPermittedIntents(allPrivilegedFlags)

	want := gateway.IntentGuilds | gateway.IntentGuildMessages | gateway.IntentGuildMessageReactions |
		gateway.IntentMessageContent | gateway.IntentGuildMembers

	if got != want {
		t.Errorf("allPermittedIntents = %d, want %d (missing %d)", got, want, want&^got)
	}
}

// Privileged intents are still gated on the portal flags, or Discord rejects
// the identify with close code 4014.
func TestAllPermittedIntentsRespectsFlags(t *testing.T) {
	got := allPermittedIntents(0)

	if got&gateway.IntentMessageContent != 0 {
		t.Error("message content granted without the portal flag")
	}
	if got&gateway.IntentGuildMembers != 0 {
		t.Error("guild members granted without the portal flag")
	}
}
