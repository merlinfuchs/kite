package gateway

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
)

func (g *Gateway) Guilds(ctx context.Context) ([]discord.Guild, error) {
	guilds, err := g.session.GuildStore.Guilds()
	if err != nil {
		return nil, err
	}

	return guilds, nil
}

func (g *Gateway) GuildChannels(ctx context.Context, guildID string) ([]discord.Channel, error) {
	gid, _ := discord.ParseSnowflake(guildID)

	channels, err := g.session.ChannelStore.Channels(discord.GuildID(gid))
	if err != nil {
		return nil, err
	}

	return channels, nil
}
