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
