package flow

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memberPruneTestProvider struct {
	provider.MockDiscordProvider

	called  bool
	guildID discord.GuildID
	data    api.PruneData
}

func (p *memberPruneTestProvider) PruneMembers(ctx context.Context, guildID discord.GuildID, data api.PruneData) (uint, error) {
	p.called = true
	p.guildID = guildID
	p.data = data
	return 5, nil
}

func executeMemberPrune(t *testing.T, p provider.DiscordProvider, data FlowNodeData) (*FlowContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{
			Discord: p,
			Log:     &provider.MockLogProvider{},
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	node := &CompiledFlowNode{ID: "1", Type: FlowNodeTypeActionMemberPrune, Data: data}
	return c, node.Execute(c)
}

func TestExecuteMemberPrune(t *testing.T) {
	p := &memberPruneTestProvider{}
	c, err := executeMemberPrune(t, p, FlowNodeData{
		GuildTarget:     "3",
		MemberPruneDays: " {{7 * 2}} ",
		AuditLogReason:  "inactive",
	})
	require.NoError(t, err)

	assert.Equal(t, discord.GuildID(3), p.guildID)
	assert.Equal(t, uint(14), p.data.Days)
	assert.True(t, p.data.ReturnCount)
	assert.Empty(t, p.data.IncludedRoles)
	assert.Equal(t, api.AuditLogReason("inactive"), p.data.AuditLogReason)
	assert.Equal(t, int64(5), c.GetNodeState("1").Result.Int())
}

func TestExecuteMemberPruneRejectsInvalidDays(t *testing.T) {
	tests := map[string]struct {
		days string
		err  string
	}{
		"empty":    {days: "", err: `prune days "" is not a whole number of days`},
		"text":     {days: "week", err: `prune days "week" is not a whole number of days`},
		"decimal":  {days: "1.5", err: `prune days "1.5" is not a whole number of days`},
		"zero":     {days: "0", err: "prune days must be between 1 and 30, got 0"},
		"negative": {days: "-7", err: "prune days must be between 1 and 30, got -7"},
		"too many": {days: "31", err: "prune days must be between 1 and 30, got 31"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := &memberPruneTestProvider{}
			_, err := executeMemberPrune(t, p, FlowNodeData{GuildTarget: "3", MemberPruneDays: tt.days})
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.err)
			assert.False(t, p.called)
		})
	}
}
