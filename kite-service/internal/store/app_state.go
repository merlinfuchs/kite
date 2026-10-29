package store

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

type AppStateStatus struct {
	Online bool
}

type AppStateManager interface {
	AppState(ctx context.Context, appID string) (AppStateStore, error)
	AppClient(ctx context.Context, appID string) (*api.Client, error)
}

type AppStateStore interface {
	AppStatus(ctx context.Context) (AppStateStatus, error)
	AppUser(ctx context.Context) (*discord.User, error)
	AppGuilds(ctx context.Context) ([]discord.Guild, error)
	AppGuildChannels(ctx context.Context, guildID string) ([]discord.Channel, error)
	AppGuildRoles(ctx context.Context, guildID string) ([]discord.Role, error)
	AppGuildEmojis(ctx context.Context, guildID string) ([]discord.Emoji, error)
	// AppDiscordAPIRequest sends a request to the Discord API with the bot
	// token and returns the response body.
	AppDiscordAPIRequest(ctx context.Context, req provider.DiscordAPIRequest) ([]byte, error)
}
