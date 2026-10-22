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

type serverLeaveDiscordProvider struct {
	provider.MockDiscordProvider

	left []discord.GuildID
}

func (p *serverLeaveDiscordProvider) LeaveGuild(ctx context.Context, guildID discord.GuildID) error {
	p.left = append(p.left, guildID)
	return nil
}

func runServerLeave(t *testing.T, contextData FlowContextData, data FlowNodeData) (*serverLeaveDiscordProvider, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discordProvider := &serverLeaveDiscordProvider{}

	c := NewContext(
		ctx,
		5*time.Second,
		contextData,
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
	p, err := runServerLeave(t, &blockTestContextData{}, FlowNodeData{})
	require.NoError(t, err)
	assert.Equal(t, []discord.GuildID{5}, p.left)
}

// A guild target in the node data must be ignored, or a flow in one server
// could make the app leave another one.
func TestServerLeaveIgnoresGuildTarget(t *testing.T) {
	p, err := runServerLeave(t, &blockTestContextData{}, FlowNodeData{GuildTarget: "222"})
	require.NoError(t, err)
	assert.Equal(t, []discord.GuildID{5}, p.left)
}

func TestServerLeaveFailsOutsideServer(t *testing.T) {
	p, err := runServerLeave(t, &TestContextData{}, FlowNodeData{})
	require.Error(t, err)
	assert.Empty(t, p.left)
}
