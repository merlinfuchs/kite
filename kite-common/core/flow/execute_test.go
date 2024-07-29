package flow

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var flowCommandTest = CompiledFlowNode{
	ID:   "0",
	Type: FlowNodeTypeEntryCommand,
	Data: FlowNodeData{
		Name:        "ping",
		Description: "Pong!",
	},
	Children: []*CompiledFlowNode{
		{
			ID:   "1",
			Type: FlowNodeTypeConditionCompare,
			Data: FlowNodeData{
				ConditionBaseValue: FlowValueNull,
			},
			Children: []*CompiledFlowNode{
				{
					ID:   "2",
					Type: FlowNodeTypeConditionItemCompare,
					Data: FlowNodeData{
						ConditionItemMode:  ConditionItemModeEqual,
						ConditionItemValue: FlowValueNull,
					},
					Children: []*CompiledFlowNode{
						{
							ID:   "3",
							Type: FlowNodeTypeActionResponseCreate,
							Data: FlowNodeData{
								MessageData: api.SendMessageData{
									Content: "Pong!",
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestFlowExecuteCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discordProvider := &TestDiscordProvider{}

	c := NewContext(
		ctx,
		&TestContextData{},
		FlowProviders{
			Discord: discordProvider,
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxActions:    1,
		},
		nil,
	)

	err := flowCommandTest.Execute(c)
	require.NoError(t, err)
	assert.Equal(t, "Pong!", discordProvider.response.Data.Content.Val)
}

type TestDiscordProvider struct {
	response api.InteractionResponse
}

func (p *TestDiscordProvider) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) error {
	p.response = response
	return nil
}

func (p *TestDiscordProvider) CreateMessage(ctx context.Context, channelID discord.ChannelID, message api.SendMessageData) (*discord.Message, error) {
	return nil, nil
}

type TestContextData struct{}

func (d *TestContextData) Interaction() *discord.InteractionEvent {
	return &discord.InteractionEvent{}
}

func (d *TestContextData) GuildID() discord.GuildID {
	return 0
}

func (d *TestContextData) ChannelID() discord.ChannelID {
	return 0
}

func (d *TestContextData) CommandData() *discord.CommandInteraction {
	return nil
}

func (d *TestContextData) MessageComponentData() discord.ComponentInteraction {
	return nil
}

func (d *TestContextData) EventData() gateway.Event {
	return nil
}
