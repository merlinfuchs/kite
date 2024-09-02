package store

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
)

type AppStateManager interface {
	AppState(ctx context.Context, appID string) (AppStateStore, error)
}

type AppStateStore interface {
	Guilds(ctx context.Context) ([]discord.Guild, error)
	GuildChannels(ctx context.Context, guildID string) ([]discord.Channel, error)
}
