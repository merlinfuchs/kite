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

type presenceTestProvider struct {
	provider.MockDiscordProvider
	presences map[discord.UserID]discord.Presence
}

func (p *presenceTestProvider) MemberPresence(ctx context.Context, guildID discord.GuildID, userID discord.UserID) (*discord.Presence, error) {
	presence, ok := p.presences[userID]
	if !ok {
		return nil, provider.ErrNotFound
	}
	return &presence, nil
}

func TestMemberPresenceGet(t *testing.T) {
	discordProvider := &presenceTestProvider{presences: map[discord.UserID]discord.Presence{
		1: {Status: discord.DoNotDisturbStatus, Activities: []discord.Activity{
			{Name: "Minecraft", Type: discord.GameActivity},
			{Name: "Custom Status", Type: discord.CustomActivity, State: "sub to jack", Emoji: &discord.Emoji{Name: "🔥"}},
		}},
		2: {Status: discord.InvisibleStatus},
	}}

	get := func(userID string) string {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		t.Cleanup(cancel)

		logs := &recordingLogProvider{}
		c := NewContext(
			ctx,
			5*time.Second,
			&TestContextData{},
			FlowProviders{Discord: discordProvider, Log: logs},
			FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
			eval.NewContext(eval.Env{}),
			nil,
		)
		t.Cleanup(c.Cancel)

		node := &CompiledFlowNode{
			ID:   "presence_abc",
			Type: FlowNodeTypeActionMemberPresenceGet,
			Data: FlowNodeData{GuildTarget: "100", UserTarget: userID},
			Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{{
				ID:   "log_def",
				Type: FlowNodeTypeActionLog,
				Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "{{result('presence_abc').status}}|{{result('presence_abc').custom_status}}|{{result('presence_abc').custom_status_emoji}}|{{result('presence_abc').activity}}"},
			}}},
		}
		require.NoError(t, node.Execute(c))
		require.Len(t, logs.messages, 1)
		return logs.messages[0]
	}

	assert.Equal(t, "dnd|sub to jack|🔥|Minecraft", get("1"))
	// Invisible members appear offline.
	assert.Equal(t, "offline|||", get("2"))
	// Discord only sends the presences of members who are online.
	assert.Equal(t, "offline|||", get("3"))
}
