package store

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

type AppStateStatus struct {
	Online bool
}

type AppStateManager interface {
	AppState(ctx context.Context, appID string) (AppStateStore, error)
	AppClient(ctx context.Context, appID string) (*state.State, error)
}

type AppStateStore interface {
	AppStatus(ctx context.Context) (AppStateStatus, error)
	AppGuilds(ctx context.Context) ([]discord.Guild, error)
	AppGuildChannels(ctx context.Context, guildID string) ([]discord.Channel, error)
}
