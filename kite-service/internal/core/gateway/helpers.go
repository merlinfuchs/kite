package gateway

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

const (
	GATEWAY_GUILD_MEMBERS           = 1 << 14
	GATEWAY_GUILD_MEMBERS_LIMITED   = 1 << 15
	GATEWAY_MESSAGE_CONTENT         = 1 << 18
	GATEWAY_MESSAGE_CONTENT_LIMITED = 1 << 19
)

// allPermittedIntents returns every intent the app is approved for. Used when
// requirements cannot be loaded, so a database blip degrades to the old
// unconditional behaviour rather than to dropping events.
func allPermittedIntents(flags discord.ApplicationFlags) gateway.Intents {
	res := gateway.IntentGuilds | gateway.IntentGuildMessages | gateway.IntentGuildMessageReactions

	if flags&GATEWAY_MESSAGE_CONTENT != 0 || flags&GATEWAY_MESSAGE_CONTENT_LIMITED != 0 {
		res |= gateway.IntentMessageContent
	}
	if flags&GATEWAY_GUILD_MEMBERS != 0 || flags&GATEWAY_GUILD_MEMBERS_LIMITED != 0 {
		res |= gateway.IntentGuildMembers
	}

	return res
}

// intentsForRequirements derives the smallest intent set that still delivers
// everything the app consumes.
//
// Privileged intents are gated on the app's portal flags as well as on need:
// identifying with a privileged intent the app was never approved for is
// rejected by Discord with a 4014 close code.
func intentsForRequirements(reqs model.AppGatewayRequirements, flags discord.ApplicationFlags) gateway.Intents {
	// Interactions are delivered regardless of intents, so a command-only app
	// needs nothing beyond IntentGuilds -- which is kept for everyone because
	// the dashboard's guild and channel pickers read from the state cache it
	// populates.
	res := gateway.IntentGuilds

	if reqs.NeedsGuildMessages() {
		res |= gateway.IntentGuildMessages

		if flags&GATEWAY_MESSAGE_CONTENT != 0 || flags&GATEWAY_MESSAGE_CONTENT_LIMITED != 0 {
			res |= gateway.IntentMessageContent
		}
	}

	if reqs.NeedsGuildMembers() {
		if flags&GATEWAY_GUILD_MEMBERS != 0 || flags&GATEWAY_GUILD_MEMBERS_LIMITED != 0 {
			res |= gateway.IntentGuildMembers
		}
	}

	if reqs.NeedsGuildMessageReactions() {
		res |= gateway.IntentGuildMessageReactions
	}

	return res
}

func createSession(tokenCrypt *util.SymmetricCrypt, app *model.App) (*state.State, error) {
	token, err := tokenCrypt.DecryptString(app.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	identifier := gateway.DefaultIdentifier("Bot " + token)
	identifier.IdentifyCommand.Presence = presenceForApp(app)

	// TODO: pass in custom opts instead of modifying the default
	gateway.DefaultGatewayOpts.AlwaysCloseGracefully = false

	// TODO: configure state to only cache what we need
	return state.NewWithIdentifier(identifier), nil
}

func presenceForApp(app *model.App) *gateway.UpdatePresenceCommand {
	var entry *model.AppDiscordStatusEntry
	if app.DiscordStatus != nil {
		entry = app.DiscordStatus.ActiveEntry()
	}

	return presenceForStatusEntry(entry)
}

// presenceForStatusEntry falls back to Kite's default presence when entry is nil.
func presenceForStatusEntry(entry *model.AppDiscordStatusEntry) *gateway.UpdatePresenceCommand {
	status := discord.OnlineStatus
	activity := discord.Activity{
		Type:  discord.CustomActivity,
		Name:  "kite.onl",
		State: "🪁 Powered by Kite.onl",
	}

	if entry != nil {
		if entry.Status != "" {
			status = discord.Status(entry.Status)
		}

		activity = discord.Activity{
			Type:  discord.ActivityType(entry.ActivityType),
			Name:  entry.ActivityName,
			State: entry.ActivityState,
			URL:   entry.ActivityURL,
		}
	}

	return &gateway.UpdatePresenceCommand{
		Status:     status,
		Activities: []discord.Activity{activity},
	}
}
