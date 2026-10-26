package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	arikawastore "github.com/diamondburned/arikawa/v3/state/store"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

func (g *Gateway) AppStatus(ctx context.Context) (store.AppStateStatus, error) {
	return store.AppStateStatus{
		Online: g.Session().GatewayIsAlive(),
	}, nil
}

// AppUser returns the bot's user, which the session keeps from the Ready event
// and user updates, so it doesn't need a request to Discord.
func (g *Gateway) AppUser(ctx context.Context) (*discord.User, error) {
	return g.Session().WithContext(ctx).Me()
}

func (g *Gateway) AppGuilds(ctx context.Context) ([]discord.Guild, error) {
	guilds, err := g.Session().GuildStore.Guilds()
	if err != nil {
		return nil, err
	}

	return guilds, nil
}

// AppGuildChannels returns no channels for a guild the bot isn't in. Channels
// that aren't cached yet, like while the bot reconnects, are fetched from
// Discord.
func (g *Gateway) AppGuildChannels(ctx context.Context, guildID string) ([]discord.Channel, error) {
	gid, err := discord.ParseSnowflake(guildID)
	if err != nil {
		return nil, fmt.Errorf("invalid guild ID: %w", err)
	}

	session := g.Session()
	if g.knownNotInGuild(session, discord.GuildID(gid)) {
		return nil, nil
	}

	channels, err := session.WithContext(ctx).Channels(discord.GuildID(gid))
	if isNotInGuild(err) {
		g.rememberNotInGuild(discord.GuildID(gid))
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

	session := g.Session()
	if g.knownNotInGuild(session, discord.GuildID(gid)) {
		return nil, nil
	}

	roles, err := session.WithContext(ctx).Roles(discord.GuildID(gid))
	if isNotInGuild(err) {
		g.rememberNotInGuild(discord.GuildID(gid))
		return nil, nil
	}
	return roles, err
}

const (
	// notInGuildTTL is how long a guild Discord said the bot isn't in isn't
	// asked about again. Discord blocks Kite's IP for all bots after too many
	// failed requests, which repeated lookups of such a guild would add to.
	notInGuildTTL = 10 * time.Minute
	// maxNotInGuild bounds how many such guilds are remembered.
	maxNotInGuild = 1000
)

// knownNotInGuild says whether Discord recently said the bot isn't in the
// guild. A guild in the cache was joined since, so it doesn't count.
func (g *Gateway) knownNotInGuild(session *state.State, guildID discord.GuildID) bool {
	if _, err := session.GuildStore.Guild(guildID); err == nil {
		return false
	}

	g.notInGuildMu.Lock()
	defer g.notInGuildMu.Unlock()
	until, ok := g.notInGuild[guildID]
	return ok && time.Now().Before(until)
}

func (g *Gateway) rememberNotInGuild(guildID discord.GuildID) {
	g.notInGuildMu.Lock()
	defer g.notInGuildMu.Unlock()

	now := time.Now()
	if len(g.notInGuild) >= maxNotInGuild {
		for id, until := range g.notInGuild {
			if now.After(until) {
				delete(g.notInGuild, id)
			}
		}
	}
	if g.notInGuild == nil || len(g.notInGuild) >= maxNotInGuild {
		g.notInGuild = make(map[discord.GuildID]time.Time)
	}
	g.notInGuild[guildID] = now.Add(notInGuildTTL)
}

// isNotInGuild reports whether Discord rejected a request because the bot isn't
// in the guild or the guild doesn't exist.
func isNotInGuild(err error) bool {
	return util.IsDiscordRestStatusCode(err, http.StatusForbidden) ||
		util.IsDiscordRestStatusCode(err, http.StatusNotFound)
}

// AppGuildEmojis returns the custom emojis of a guild the app is in. They come
// from the cache, which GUILD_CREATE fills, so no request is sent to Discord.
func (g *Gateway) AppGuildEmojis(ctx context.Context, guildID string) ([]discord.Emoji, error) {
	gid, err := discord.ParseSnowflake(guildID)
	if err != nil {
		return nil, fmt.Errorf("invalid guild ID: %w", err)
	}

	emojis, err := g.Session().EmojiStore.Emojis(discord.GuildID(gid))
	if err != nil {
		if errors.Is(err, arikawastore.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return emojis, nil
}
