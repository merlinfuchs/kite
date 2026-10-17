package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
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

// AppGuildChannels returns no channels for a guild the bot isn't in. Channels
// that aren't cached yet, like right after connecting, are fetched from
// Discord.
func (g *Gateway) AppGuildChannels(ctx context.Context, guildID string) ([]discord.Channel, error) {
	gid, err := discord.ParseSnowflake(guildID)
	if err != nil {
		return nil, fmt.Errorf("invalid guild ID: %w", err)
	}

	channels, err := g.Session().WithContext(ctx).Channels(discord.GuildID(gid))
	if isNotInGuild(err) {
		return nil, nil
	}
	return channels, err
}

// AppGuildRoles returns no roles for a guild the bot isn't in, and fetches
// roles that aren't cached yet like AppGuildChannels.
func (g *Gateway) AppGuildRoles(ctx context.Context, guildID string) ([]discord.Role, error) {
	gid, err := discord.ParseSnowflake(guildID)
	if err != nil {
		return nil, fmt.Errorf("invalid guild ID: %w", err)
	}

	roles, err := g.Session().WithContext(ctx).Roles(discord.GuildID(gid))
	if isNotInGuild(err) {
		return nil, nil
	}
	return roles, err
}

// isNotInGuild reports whether Discord rejected a request because the bot isn't
// in the guild or the guild doesn't exist.
func isNotInGuild(err error) bool {
	return util.IsDiscordRestStatusCode(err, http.StatusForbidden) ||
		util.IsDiscordRestStatusCode(err, http.StatusNotFound)
}
