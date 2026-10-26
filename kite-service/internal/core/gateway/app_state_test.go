package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

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

// A guild the bot isn't in must not be asked about: the session's token is
// fake, so a request to Discord would fail with an error.
func TestAppGuildChannelsAndRolesOnlyAskForCachedGuilds(t *testing.T) {
	g := newTestGateway(t)
	ctx := context.Background()

	channels, err := g.AppGuildChannels(ctx, "123")
	assert.NoError(t, err)
	assert.Empty(t, channels)

	roles, err := g.AppGuildRoles(ctx, "123")
	assert.NoError(t, err)
	assert.Empty(t, roles)

	guildID := discord.GuildID(456)
	cabinet := g.Session().Cabinet
	assert.NoError(t, cabinet.GuildSet(&discord.Guild{ID: guildID}, false))
	assert.NoError(t, cabinet.ChannelSet(&discord.Channel{ID: 2, GuildID: guildID}, false))
	assert.NoError(t, cabinet.RoleSet(guildID, &discord.Role{ID: 1}, false))

	channels, err = g.AppGuildChannels(ctx, guildID.String())
	assert.NoError(t, err)
	assert.Len(t, channels, 1)

	roles, err = g.AppGuildRoles(ctx, guildID.String())
	assert.NoError(t, err)
	assert.Len(t, roles, 1)
}
