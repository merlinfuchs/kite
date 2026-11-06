package flow

import (
	"context"
	"strings"
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

func TestMessageCreateColorPlaceholder(t *testing.T) {
	send := func(color string) (*messageTestProvider, error) {
		p := &messageTestProvider{}
		_, err := executeBlock(t, p, FlowNodeTypeActionMessageCreate, `{
			"channel_target": "3",
			"message_data": {"embeds": [{"description": "Hi", "color": "`+color+`"}]}
		}`)
		return p, err
	}

	p, err := send(`{{'#ff0000'}}`)
	require.NoError(t, err)
	require.Len(t, p.sent.Embeds, 1)
	assert.Equal(t, discord.Color(0xFF0000), p.sent.Embeds[0].Color)

	// Empty, like the color of a role without one.
	p, err = send(`{{''}}`)
	require.NoError(t, err)
	assert.Equal(t, discord.Color(0), p.sent.Embeds[0].Color)

	_, err = send(`{{'red'}}`)
	assert.ErrorContains(t, err, "must be a number or a hex color")
}

type personaTestProvider struct {
	provider.MockDiscordProvider

	channelID discord.ChannelID
	persona   provider.Persona
	sent      api.SendMessageData
}

func (p *personaTestProvider) CreatePersonaMessage(ctx context.Context, channelID discord.ChannelID, persona provider.Persona, message api.SendMessageData) (*discord.Message, error) {
	p.channelID = channelID
	p.persona = persona
	p.sent = message
	return &discord.Message{ID: 9, ChannelID: channelID}, nil
}

func TestPersonaMessageCreate(t *testing.T) {
	p := &personaTestProvider{}
	_, err := executeBlock(t, p, FlowNodeTypeActionPersonaMessageCreate, `{
		"channel_target": "3",
		"persona_name": " Narrator ",
		"persona_avatar_url": "https://example.com/a.png",
		"message_data": {"content": "Hi"}
	}`)
	require.NoError(t, err)
	assert.Equal(t, discord.ChannelID(3), p.channelID)
	assert.Equal(t, provider.Persona{Name: "Narrator", AvatarURL: "https://example.com/a.png"}, p.persona)
	assert.Equal(t, "Hi", p.sent.Content)

	for data, want := range map[string]string{
		`"persona_name": "{{''}}"`:                                   "name is empty",
		`"persona_name": "` + strings.Repeat("a", 81) + `"`:          "longer than 80 characters",
		`"persona_name": "A", "persona_avatar_url": "ftp://x/a.png"`: "must start with http:// or https://",
	} {
		_, err := executeBlock(t, &personaTestProvider{}, FlowNodeTypeActionPersonaMessageCreate, `{
			"channel_target": "3",
			`+data+`,
			"message_data": {"content": "Hi"}
		}`)
		assert.ErrorContains(t, err, want, data)
	}
}
