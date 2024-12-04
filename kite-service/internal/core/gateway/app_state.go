package gateway

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (g *Gateway) AppStatus(ctx context.Context) (store.AppStateStatus, error) {
	return store.AppStateStatus{
		Online: g.session != nil && g.session.GatewayIsAlive(),
	}, nil
}

func (g *Gateway) AppGuilds(ctx context.Context) ([]discord.Guild, error) {
	guilds, err := g.session.GuildStore.Guilds()
	if err != nil {
		return nil, err
	}

	return guilds, nil
}

func (g *Gateway) AppGuildChannels(ctx context.Context, guildID string) ([]discord.Channel, error) {
	gid, _ := discord.ParseSnowflake(guildID)

	channels, err := g.session.ChannelStore.Channels(discord.GuildID(gid))
	if err != nil {
		return nil, err
	}

	return channels, nil
}
