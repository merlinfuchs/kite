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

// AppGuildChannelPermissions returns the bot's own permissions in each of the
// passed channels of a guild, after its roles and the channels' overwrites.
func (g *Gateway) AppGuildChannelPermissions(ctx context.Context, guildID string, channels []discord.Channel) (map[discord.ChannelID]discord.Permissions, error) {
	gid, err := discord.ParseSnowflake(guildID)
	if err != nil {
		return nil, fmt.Errorf("invalid guild ID: %w", err)
	}

	session := g.Session().WithContext(ctx)

	me, err := session.Me()
	if err != nil {
		return nil, fmt.Errorf("failed to get bot user: %w", err)
	}

	guild, err := session.Guild(discord.GuildID(gid))
	if err != nil {
		return nil, fmt.Errorf("failed to get guild: %w", err)
	}

	member, err := session.Member(discord.GuildID(gid), me.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot member: %w", err)
	}

	roles, err := session.Roles(discord.GuildID(gid))
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	byID := make(map[discord.ChannelID]discord.Channel, len(channels))
	for _, channel := range channels {
		byID[channel.ID] = channel
	}

	res := make(map[discord.ChannelID]discord.Permissions, len(channels))
	for _, channel := range channels {
		// Threads have no overwrites of their own and use their parent's.
		target := channel
		if isThread(channel.Type) {
			if parent, ok := byID[channel.ParentID]; ok {
				target = parent
			}
		}

		res[channel.ID] = discord.CalcOverrides(*guild, target, *member, roles)
	}

	return res, nil
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

// AppGuildMembers returns up to limit members of a guild, always fetched from
// Discord because the member cache only holds members the bot has seen.
// Listing members needs the Server Members intent, so Discord rejects the
// request with 403 when it's turned off, which is returned as is. A guild that
// doesn't exist returns no members.
func (g *Gateway) AppGuildMembers(ctx context.Context, guildID string, limit uint) ([]discord.Member, error) {
	gid, err := discord.ParseSnowflake(guildID)
	if err != nil {
		return nil, fmt.Errorf("invalid guild ID: %w", err)
	}

	members, err := g.Session().Client.WithContext(ctx).Members(discord.GuildID(gid), limit)
	if util.IsDiscordRestStatusCode(err, http.StatusNotFound) {
		return nil, nil
	}
	return members, err
}

func isThread(t discord.ChannelType) bool {
	return t == discord.GuildAnnouncementThread ||
		t == discord.GuildPublicThread ||
		t == discord.GuildPrivateThread
}

// isNotInGuild reports whether Discord rejected a request because the bot isn't
// in the guild or the guild doesn't exist.
func isNotInGuild(err error) bool {
	return util.IsDiscordRestStatusCode(err, http.StatusForbidden) ||
		util.IsDiscordRestStatusCode(err, http.StatusNotFound)
}
