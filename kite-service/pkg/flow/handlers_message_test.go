package flow

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
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
		"message_target": "5",
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
		"message_target": "{{missing}}",
		"message_data": {"content": "Hi"}
	}`)
	assert.ErrorContains(t, err, "message to reply to must be an ID")
}
