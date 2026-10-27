package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/stretchr/testify/assert"
)

func TestIsNotInGuild(t *testing.T) {
	assert.True(t, isNotInGuild(&httputil.HTTPError{Status: http.StatusForbidden}))
	assert.True(t, isNotInGuild(fmt.Errorf("wrapped: %w", &httputil.HTTPError{Status: http.StatusNotFound})))

	assert.False(t, isNotInGuild(nil))
	assert.False(t, isNotInGuild(&httputil.HTTPError{Status: http.StatusInternalServerError}))
	assert.False(t, isNotInGuild(errors.New("timeout")))
}

// Guilds Discord said the bot isn't in aren't asked about again. The
// session's token is fake, so a request to Discord would fail with an error.
func TestAppGuildChannelsAndRolesRememberGuildsTheBotIsNotIn(t *testing.T) {
	g := newTestGateway(t)
	ctx := context.Background()

	g.rememberNotInGuild(123)
	channels, err := g.AppGuildChannels(ctx, "123")
	assert.NoError(t, err)
	assert.Empty(t, channels)
	roles, err := g.AppGuildRoles(ctx, "123")
	assert.NoError(t, err)
	assert.Empty(t, roles)

	// A guild the bot joined since is in the cache, which wins.
	cabinet := g.Session().Cabinet
	assert.NoError(t, cabinet.GuildSet(&discord.Guild{ID: 123}, false))
	assert.NoError(t, cabinet.ChannelSet(&discord.Channel{ID: 2, GuildID: 123}, false))
	assert.NoError(t, cabinet.RoleSet(123, &discord.Role{ID: 1}, false))
	channels, err = g.AppGuildChannels(ctx, "123")
	assert.NoError(t, err)
	assert.Len(t, channels, 1)
	roles, err = g.AppGuildRoles(ctx, "123")
	assert.NoError(t, err)
	assert.Len(t, roles, 1)
}

func TestNotInGuildExpiresAndIsBounded(t *testing.T) {
	g := newTestGateway(t)
	session := g.Session()

	g.rememberNotInGuild(1)
	assert.True(t, g.knownNotInGuild(session, 1))
	g.notInGuild[1] = time.Now().Add(-time.Second)
	assert.False(t, g.knownNotInGuild(session, 1))

	for i := range maxNotInGuild + 10 {
		g.rememberNotInGuild(discord.GuildID(i + 2))
	}
	assert.LessOrEqual(t, len(g.notInGuild), maxNotInGuild)
}

func TestAppGuildBotPermissions(t *testing.T) {
	g := newTestGateway(t)
	ctx := context.Background()
	cabinet := g.Session().Cabinet

	botUser := discord.User{ID: 999, Username: "KiteBot"}
	assert.NoError(t, cabinet.MyselfSet(botUser, false))

	// 1. Bot is server owner
	guildOwner := discord.Guild{ID: 10, OwnerID: 999}
	assert.NoError(t, cabinet.GuildSet(&guildOwner, false))
	perms, err := g.AppGuildBotPermissions(ctx, "10")
	assert.NoError(t, err)
	assert.Equal(t, ^discord.Permissions(0), perms)

	// 2. Normal permissions from @everyone and member roles
	guildNormal := discord.Guild{ID: 20, OwnerID: 111}
	assert.NoError(t, cabinet.GuildSet(&guildNormal, false))
	// @everyone role
	assert.NoError(t, cabinet.RoleSet(20, &discord.Role{
		ID:          20,
		Permissions: discord.PermissionViewChannel | discord.PermissionSendMessages,
	}, false))
	// Custom role
	assert.NoError(t, cabinet.RoleSet(20, &discord.Role{
		ID:          201,
		Permissions: discord.PermissionBanMembers,
	}, false))
	// Member
	assert.NoError(t, cabinet.MemberSet(20, &discord.Member{
		User:    botUser,
		RoleIDs: []discord.RoleID{201},
	}, false))

	perms, err = g.AppGuildBotPermissions(ctx, "20")
	assert.NoError(t, err)
	expectedPerms := discord.PermissionViewChannel | discord.PermissionSendMessages | discord.PermissionBanMembers
	assert.Equal(t, expectedPerms, perms)

	// 3. Admin role gives all permissions
	assert.NoError(t, cabinet.RoleSet(20, &discord.Role{
		ID:          202,
		Permissions: discord.PermissionAdministrator,
	}, false))
	assert.NoError(t, cabinet.MemberSet(20, &discord.Member{
		User:    botUser,
		RoleIDs: []discord.RoleID{201, 202},
	}, true))

	perms, err = g.AppGuildBotPermissions(ctx, "20")
	assert.NoError(t, err)
	assert.Equal(t, ^discord.Permissions(0), perms)

	// 4. Baseline @everyone only (member has no roles assigned)
	guildBaseline := discord.Guild{ID: 30, OwnerID: 111}
	assert.NoError(t, cabinet.GuildSet(&guildBaseline, false))
	assert.NoError(t, cabinet.RoleSet(30, &discord.Role{
		ID:          30,
		Permissions: discord.PermissionViewChannel,
	}, false))
	assert.NoError(t, cabinet.MemberSet(30, &discord.Member{
		User:    botUser,
		RoleIDs: []discord.RoleID{},
	}, false))
	perms, err = g.AppGuildBotPermissions(ctx, "30")
	assert.NoError(t, err)
	assert.Equal(t, discord.PermissionViewChannel, perms)

	// 5. Invalid snowflake guild ID
	_, err = g.AppGuildBotPermissions(ctx, "not-a-snowflake")
	assert.Error(t, err)

	// 6. Guild marked in notInGuild returns 0, nil
	g.rememberNotInGuild(40)
	perms, err = g.AppGuildBotPermissions(ctx, "40")
	assert.NoError(t, err)
	assert.Equal(t, discord.Permissions(0), perms)
}
