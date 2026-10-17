package flow

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serverLeaveContextData struct {
	TestContextData
	guildID discord.GuildID
}

func (d *serverLeaveContextData) GuildID() discord.GuildID {
	return d.guildID
}

type serverLeaveDiscordProvider struct {
	provider.MockDiscordProvider

	left []discord.GuildID
}

func (p *serverLeaveDiscordProvider) LeaveGuild(ctx context.Context, guildID discord.GuildID) error {
	p.left = append(p.left, guildID)
	return nil
}

func runServerLeave(t *testing.T, guildID discord.GuildID, data FlowNodeData) (*serverLeaveDiscordProvider, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discordProvider := &serverLeaveDiscordProvider{}

	c := NewContext(
		ctx,
		5*time.Second,
		&serverLeaveContextData{guildID: guildID},
		FlowProviders{
			Discord: discordProvider,
			Log:     &provider.MockLogProvider{},
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	defer c.Cancel()

	node := CompiledFlowNode{
		ID:   "0",
		Type: FlowNodeTypeActionServerLeave,
		Data: data,
	}

	return discordProvider, node.Execute(c)
}

func TestServerLeaveLeavesCurrentServer(t *testing.T) {
	p, err := runServerLeave(t, 111, FlowNodeData{})
	require.NoError(t, err)
	assert.Equal(t, []discord.GuildID{111}, p.left)
}

// Flows saved before the server field was removed may still carry a guild
// target. It must be ignored, or a flow in one server could make the app leave
// another one.
func TestServerLeaveIgnoresGuildTarget(t *testing.T) {
	p, err := runServerLeave(t, 111, FlowNodeData{GuildTarget: "222"})
	require.NoError(t, err)
	assert.Equal(t, []discord.GuildID{111}, p.left)
}

func TestServerLeaveFailsOutsideServer(t *testing.T) {
	p, err := runServerLeave(t, 0, FlowNodeData{})
	require.Error(t, err)
	assert.Empty(t, p.left)
}
