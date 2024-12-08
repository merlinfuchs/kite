package gateway

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

const GATEWAY_GUILD_MEMBERS = 1 << 14
const GATEWAY_GUILD_MEMBERS_LIMITED = 1 << 15
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
	if app.Flags&GATEWAY_GUILD_MEMBERS != 0 || app.Flags&GATEWAY_GUILD_MEMBERS_LIMITED != 0 {
		res |= gateway.IntentGuildMembers
	}

	return res, nil
}

func createSession(app *model.App) *state.State {
	identifier := gateway.DefaultIdentifier("Bot " + app.DiscordToken)
	identifier.IdentifyCommand.Presence = presenceForApp(app)

	// TODO: pass in custom opts instead of modifying the default
	gateway.DefaultGatewayOpts.AlwaysCloseGracefully = false

	// TODO: configure state to only cache what we need
	return state.NewWithIdentifier(identifier)
}

func presenceForApp(app *model.App) *gateway.UpdatePresenceCommand {
	status := discord.OnlineStatus
	activity := discord.Activity{
		Type:  discord.CustomActivity,
		Name:  "kite.onl",
		State: "🪁 Powered by Kite.onl",
	}

	if app.DiscordStatus != nil {
		if app.DiscordStatus.Status != "" {
			status = discord.Status(app.DiscordStatus.Status)
		}

		activity = discord.Activity{
			Type:  discord.ActivityType(app.DiscordStatus.ActivityType),
			Name:  app.DiscordStatus.ActivityName,
			State: app.DiscordStatus.ActivityState,
			URL:   app.DiscordStatus.ActivityURL,
		}
	}

	return &gateway.UpdatePresenceCommand{
		Status:     status,
		Activities: []discord.Activity{activity},
	}
}
