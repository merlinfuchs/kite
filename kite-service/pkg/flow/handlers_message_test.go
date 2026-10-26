package flow

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type messageTestProvider struct {
	provider.MockDiscordProvider

	sent api.SendMessageData
}

func (p *messageTestProvider) CreateMessage(ctx context.Context, channelID discord.ChannelID, message api.SendMessageData) (*discord.Message, error) {
	p.sent = message
	return &discord.Message{ID: 9, ChannelID: channelID}, nil
}

func TestMessageCreateReply(t *testing.T) {
	p := &messageTestProvider{}
	_, err := executeBlock(t, p, FlowNodeTypeActionMessageCreate, `{
		"channel_target": "3",
		"reply_target": "5",
		"message_data": {"content": "Hi"}
	}`)
	require.NoError(t, err)
	require.NotNil(t, p.sent.Reference)
	assert.Equal(t, discord.MessageID(5), p.sent.Reference.MessageID)

	p = &messageTestProvider{}
	_, err = executeBlock(t, p, FlowNodeTypeActionMessageCreate, `{"channel_target": "3", "message_data": {"content": "Hi"}}`)
	require.NoError(t, err)
	assert.Nil(t, p.sent.Reference)

	p = &messageTestProvider{}
	_, err = executeBlock(t, p, FlowNodeTypeActionMessageCreate, `{
		"channel_target": "3",
		"reply_target": "{{missing}}",
		"message_data": {"content": "Hi"}
	}`)
	assert.ErrorContains(t, err, "message to reply to must be an ID")
}

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
	p := &privateMessageTestProvider{}
	c, err := executeBlock(t, p, FlowNodeTypeActionPrivateMessageEdit, `{
		"user_target": "5",
		"message_target": "9",
		"message_data": {"content": "Updated"}
	}`)
	require.NoError(t, err)

	// The message is edited in the DM channel of the user.
	assert.Equal(t, discord.UserID(5), p.userID)
	assert.Equal(t, discord.ChannelID(7), p.channelID)
	assert.Equal(t, discord.MessageID(9), p.messageID)
	assert.Equal(t, "Updated", p.edit.Content.Val)

	result := c.GetNodeResult("1")
	assert.Equal(t, thing.TypeDiscordMessage, result.Type)
	assert.Equal(t, discord.MessageID(9), result.DiscordMessage().ID)
}
