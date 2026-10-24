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

type categoryTestProvider struct {
	provider.MockDiscordProvider

	channels map[discord.ChannelID]discord.Channel
}

func (p *categoryTestProvider) Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error) {
	channel, ok := p.channels[channelID]
	if !ok {
		return nil, provider.ErrNotFound
	}
	return &channel, nil
}

// executeCategoryCondition runs a category condition on baseValue, with a
// branch per category in items that logs its mode and ID, and an else that
// logs "else". The flow runs in channel 6.
func executeCategoryCondition(t *testing.T, baseValue string, allowMultiple bool, items ...[2]string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	discordProvider := &categoryTestProvider{channels: map[discord.ChannelID]discord.Channel{
		// Category 10 holds text channels 6 and 7, and 7 has thread 8.
		10: {ID: 10, Type: discord.GuildCategory},
		6:  {ID: 6, Type: discord.GuildText, ParentID: 10},
		7:  {ID: 7, Type: discord.GuildText, ParentID: 10},
		8:  {ID: 8, Type: discord.GuildPublicThread, ParentID: 7},
		// Category 20 holds channel 21, and 30 isn't in a category.
		20: {ID: 20, Type: discord.GuildCategory},
		21: {ID: 21, Type: discord.GuildText, ParentID: 20},
		30: {ID: 30, Type: discord.GuildText},
	}}

	logs := &recordingLogProvider{}
	c := NewContext(
		ctx,
		5*time.Second,
		&blockTestContextData{},
		FlowProviders{Discord: discordProvider, Log: logs},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	logNode := func(id string, message string) *CompiledFlowNode {
		return &CompiledFlowNode{ID: id + "-log", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: message}}
	}

	condition := &CompiledFlowNode{
		ID:   "condition",
		Type: FlowNodeTypeControlConditionCategory,
		Data: FlowNodeData{ConditionBaseValue: baseValue, ConditionAllowMultiple: allowMultiple},
	}
	children := []*CompiledFlowNode{}
	for i, item := range items {
		id := string(rune('a' + i))
		child := &CompiledFlowNode{
			ID:   id,
			Type: FlowNodeTypeControlConditionItemCategory,
			Data: FlowNodeData{ConditionItemMode: ComparsionMode(item[0]), ConditionItemValue: item[1]},
		}
		child.Children.Default = []*CompiledFlowNode{logNode(id, item[0]+" "+item[1])}
		child.Parents.Default = []*CompiledFlowNode{condition}
		children = append(children, child)
	}
	elseNode := &CompiledFlowNode{ID: "else", Type: FlowNodeTypeControlConditionItemElse}
	elseNode.Children.Default = []*CompiledFlowNode{logNode("else", "else")}
	elseNode.Parents.Default = []*CompiledFlowNode{condition}
	condition.Children.Default = append(children, elseNode)

	err := condition.Execute(c)
	return logs.messages, err
}

func TestConditionCategory(t *testing.T) {
	in := func(id string) [2]string { return [2]string{"equal", id} }
	notIn := func(id string) [2]string { return [2]string{"not_equal", id} }

	tests := []struct {
		name     string
		base     string
		multiple bool
		items    [][2]string
		want     []string
	}{
		{"current channel", "", false, [][2]string{in("20"), in("10")}, []string{"equal 10"}},
		{"other channel", "21", false, [][2]string{in("10"), in("20")}, []string{"equal 20"}},
		{"thread is in category of its channel", "8", false, [][2]string{in("10")}, []string{"equal 10"}},
		{"category is in itself", "20", false, [][2]string{in("20")}, []string{"equal 20"}},
		{"no category", "30", false, [][2]string{in("10"), in("20")}, []string{"else"}},
		{"not in category", "30", false, [][2]string{notIn("10")}, []string{"not_equal 10"}},
		{"not in category of the channel", "", false, [][2]string{notIn("10")}, []string{"else"}},
		{"first match only", "", false, [][2]string{notIn("20"), in("10")}, []string{"not_equal 20"}},
		{"every match", "", true, [][2]string{notIn("20"), in("10")}, []string{"not_equal 20", "equal 10"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := executeCategoryCondition(t, tt.base, tt.multiple, tt.items...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, logs)
		})
	}
}

func TestConditionCategoryUnknownChannel(t *testing.T) {
	_, err := executeCategoryCondition(t, "99", false, [2]string{"equal", "10"})
	assert.Error(t, err)
}
