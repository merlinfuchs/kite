package gateway

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

const GATEWAY_MESSAGE_CONTENT = 18
const GATEWAY_MESSAGE_CONTENT_LIMITED = 19

func getAppIntents(client *api.Client) (gateway.Intents, error) {
	app, err := client.CurrentApplication()
	if err != nil {
		return 0, fmt.Errorf("failed to get current application: %w", err)
	}

	res := gateway.IntentGuilds | gateway.IntentGuildMessages
	if app.Flags&GATEWAY_MESSAGE_CONTENT != 0 || app.Flags&GATEWAY_MESSAGE_CONTENT_LIMITED != 0 {
		res |= gateway.IntentMessageContent
	}

	return res, nil
}

func createSession(app *model.App) *state.State {
	identifier := gateway.DefaultIdentifier("Bot " + app.DiscordToken)
	identifier.IdentifyCommand.Presence = &gateway.UpdatePresenceCommand{
		Status: discord.OnlineStatus,
		Activities: []discord.Activity{
			{
				Type:  discord.CustomActivity,
				Name:  "kite.onl",
				State: "🪁 Powered by Kite.onl",
			},
		},
	}

	// TODO: configure state to only cache what we need
	return state.NewWithIdentifier(identifier)
}
