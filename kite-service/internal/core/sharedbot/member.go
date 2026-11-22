package sharedbot

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

const (
	// memberTTL is how long a user's permissions in a server are trusted, so
	// losing Manage Server takes effect within it.
	memberTTL = time.Minute
	// maxMembers bounds the users remembered.
	maxMembers = 10_000
)

// Members asks Discord about users in the servers of the Kite bot, for who
// can change the apps on it.
type Members struct {
	bot *Bot

	mu    sync.Mutex
	cache map[memberKey]cachedMember
}

type memberKey struct {
	guildID discord.GuildID
	userID  discord.UserID
}

type cachedMember struct {
	member  model.SharedAppMember
	expires time.Time
}

// NewMembers returns nil without a Kite bot.
func NewMembers(bot *Bot) *Members {
	if bot == nil {
		return nil
	}
	return &Members{bot: bot, cache: make(map[memberKey]cachedMember)}
}

// Member returns the user of the Discord ID in the app's server.
func (m *Members) Member(ctx context.Context, app *model.App, discordUserID string) (model.SharedAppMember, error) {
	guildID, err := discord.ParseSnowflake(app.DiscordGuildID.String)
	if err != nil {
		return model.SharedAppMember{}, fmt.Errorf("invalid server of shared app: %w", err)
	}
	userID, err := discord.ParseSnowflake(discordUserID)
	if err != nil {
		return model.SharedAppMember{}, fmt.Errorf("invalid Discord user: %w", err)
	}
	key := memberKey{discord.GuildID(guildID), discord.UserID(userID)}

	now := time.Now()
	m.mu.Lock()
	cached, ok := m.cache[key]
	m.mu.Unlock()
	if ok && now.Before(cached.expires) {
		return cached.member, nil
	}

	member, err := m.fetch(ctx, key)
	if err != nil {
		return model.SharedAppMember{}, err
	}

	m.mu.Lock()
	if len(m.cache) >= maxMembers {
		clear(m.cache)
	}
	m.cache[key] = cachedMember{member: member, expires: now.Add(memberTTL)}
	m.mu.Unlock()
	return member, nil
}

func (m *Members) fetch(ctx context.Context, key memberKey) (model.SharedAppMember, error) {
	client := m.bot.Client(ctx)

	guild, err := client.Guild(key.guildID)
	if util.IsDiscordNotInGuild(err) {
		return model.SharedAppMember{}, nil
	}
	if err != nil {
		return model.SharedAppMember{}, fmt.Errorf("failed to get server: %w", err)
	}

	res := model.SharedAppMember{BotInServer: true}
	member, err := client.Member(key.guildID, key.userID)
	if util.IsDiscordNotInGuild(err) {
		// Not in the server.
		return res, nil
	}
	if err != nil {
		return model.SharedAppMember{}, fmt.Errorf("failed to get member: %w", err)
	}

	perms := discord.CalcOverrides(*guild, discord.Channel{}, *member, guild.Roles)
	res.ManagesServer = perms.Has(discord.PermissionManageGuild)
	res.ServerOwner = guild.OwnerID == key.userID
	for _, role := range guild.Roles {
		if slices.Contains(member.RoleIDs, role.ID) && role.Position > res.HighestRole {
			res.HighestRole = role.Position
		}
	}
	return res, nil
}
