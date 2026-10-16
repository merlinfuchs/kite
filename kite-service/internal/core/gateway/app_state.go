package gateway

import (
	"context"
	"errors"

	"github.com/diamondburned/arikawa/v3/discord"
	arikawastore "github.com/diamondburned/arikawa/v3/state/store"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (g *Gateway) AppStatus(ctx context.Context) (store.AppStateStatus, error) {
	return store.AppStateStatus{
		Online: g.Session().GatewayIsAlive(),
	}, nil
}

func (g *Gateway) AppGuilds(ctx context.Context) ([]discord.Guild, error) {
	guilds, err := g.Session().GuildStore.Guilds()
	if err != nil {
		return nil, err
	}

	return guilds, nil
}

func (g *Gateway) AppGuildChannels(ctx context.Context, guildID string) ([]discord.Channel, error) {
	gid, _ := discord.ParseSnowflake(guildID)

	channels, err := g.Session().ChannelStore.Channels(discord.GuildID(gid))
	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (g *Gateway) AppGuildRoles(ctx context.Context, guildID string) ([]discord.Role, error) {
	gid, _ := discord.ParseSnowflake(guildID)

	return g.Session().RoleStore.Roles(discord.GuildID(gid))
}

// AppGuildEmojis returns the custom emojis of a guild the app is in. They come
// from the cache, which GUILD_CREATE fills, so no request is sent to Discord.
func (g *Gateway) AppGuildEmojis(ctx context.Context, guildID string) ([]discord.Emoji, error) {
	gid, _ := discord.ParseSnowflake(guildID)

	emojis, err := g.Session().EmojiStore.Emojis(discord.GuildID(gid))
	if err != nil {
		if errors.Is(err, arikawastore.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return emojis, nil
}
