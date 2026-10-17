package flow

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type privateMessageTestProvider struct {
	provider.MockDiscordProvider

	userID    discord.UserID
	channelID discord.ChannelID
	messageID discord.MessageID
	edit      api.EditMessageData
}

func (p *privateMessageTestProvider) CreatePrivateChannel(ctx context.Context, userID discord.UserID) (*discord.Channel, error) {
	p.userID = userID
	return &discord.Channel{ID: 7, Type: discord.DirectMessage}, nil
}

func (p *privateMessageTestProvider) EditMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, data api.EditMessageData) (*discord.Message, error) {
	p.channelID = channelID
	p.messageID = messageID
	p.edit = data
	return &discord.Message{ID: messageID, ChannelID: channelID, Content: data.Content.Val}, nil
}

func TestPrivateMessageEdit(t *testing.T) {
	var data FlowNodeData
	require.NoError(t, json.Unmarshal([]byte(`{
		"user_target": "5",
		"message_target": "9",
		"message_data": {"content": "Updated"}
	}`), &data))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	p := &privateMessageTestProvider{}
	c := NewContext(
		ctx,
		5*time.Second,
		&blockTestContextData{},
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

	// The block resumes from a component when it's the entry, so it runs below
	// one.
	node := &CompiledFlowNode{
		ID:   "0",
		Type: FlowNodeTypeEntryEvent,
		Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{
			{ID: "1", Type: FlowNodeTypeActionPrivateMessageEdit, Data: data},
		}},
	}
	require.NoError(t, node.Execute(c))

	// The message is edited in the DM channel of the user.
	assert.Equal(t, discord.UserID(5), p.userID)
	assert.Equal(t, discord.ChannelID(7), p.channelID)
	assert.Equal(t, discord.MessageID(9), p.messageID)
	assert.Equal(t, "Updated", p.edit.Content.Val)

	result := c.GetNodeResult("1")
	assert.Equal(t, thing.TypeDiscordMessage, result.Type)
	assert.Equal(t, discord.MessageID(9), result.DiscordMessage().ID)
}
