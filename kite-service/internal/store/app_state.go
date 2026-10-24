package store

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
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
	AppGuildChannelPermissions(ctx context.Context, guildID string, channels []discord.Channel) (map[discord.ChannelID]discord.Permissions, error)
	AppGuildRoles(ctx context.Context, guildID string) ([]discord.Role, error)
	AppGuildMembers(ctx context.Context, guildID string, limit uint) ([]discord.Member, error)
}
