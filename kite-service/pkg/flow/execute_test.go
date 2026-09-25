package flow

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
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
	Children: ConnectedFlowNodes{
		Default: []*CompiledFlowNode{
			{
				ID:   "1",
				Type: FlowNodeTypeControlConditionCompare,
				Data: FlowNodeData{
					ConditionBaseValue: "null",
				},
				Children: ConnectedFlowNodes{
					Default: []*CompiledFlowNode{
						{
							ID:   "2",
							Type: FlowNodeTypeControlConditionItemCompare,
							Data: FlowNodeData{
								ConditionItemMode:  ComparsionModeEqual,
								ConditionItemValue: "null",
							},
							Children: ConnectedFlowNodes{
								Default: []*CompiledFlowNode{
									{
										ID:   "3",
										Type: FlowNodeTypeActionResponseCreate,
										Data: FlowNodeData{
											MessageData: &message.MessageData{
												Content: "Pong!",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func init() {
	flowCommandTest.Children.Default[0].Children.Default[0].Parents.Default = []*CompiledFlowNode{
		flowCommandTest.Children.Default[0],
	}
}

func TestFlowExecuteCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discordProvider := &TestDiscordProvider{}

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
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

	err := flowCommandTest.Execute(c)
	require.NoError(t, err)
	require.NotNil(t, discordProvider.response.Data)
	require.NotNil(t, discordProvider.response.Data.Content)
	assert.Equal(t, "Pong!", discordProvider.response.Data.Content.Val)
}

type TestDiscordProvider struct {
	provider.MockDiscordProvider

	responded bool
	response  api.InteractionResponse
}

func (p *TestDiscordProvider) HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error) {
	return p.responded, nil
}

func (p *TestDiscordProvider) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) (*provider.InteractionResponseResource, error) {
	p.response = response
	return nil, nil
}

type TestContextData struct {
	interaction *discord.InteractionEvent
}

func (d *TestContextData) Interaction() *discord.InteractionEvent {
	if d.interaction != nil {
		return d.interaction
	}
	return &discord.InteractionEvent{}
}

func (d *TestContextData) UserID() discord.UserID {
	return 0
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

func (d *TestContextData) Event() ws.Event {
	return &gateway.InteractionCreateEvent{}
}

func TestFlowExecuteModalEvaluatesTemplates(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discordProvider := &TestDiscordProvider{}

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{
			Discord:     discordProvider,
			Log:         &provider.MockLogProvider{},
			ResumePoint: &MockResumePointProvider{},
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
		Type: FlowNodeTypeEntryCommand,
		Children: ConnectedFlowNodes{
			Default: []*CompiledFlowNode{
				{
					ID:   "1",
					Type: FlowNodeTypeSuspendResponseModal,
					Data: FlowNodeData{
						ModalData: &ModalData{
							Title: "Form {{ 1 + 1 }}",
							Components: []ModalComponentData{{
								Components: []ModalComponentData{{
									CustomID:    "name_{{ 1 }}",
									Style:       1,
									Label:       "Label {{ 2 + 1 }}",
									Placeholder: "Placeholder {{ 4 }}",
									Value:       "Value {{ 5 }}",
								}},
							}},
						},
					},
				},
			},
		},
	}

	err := node.Execute(c)
	require.NoError(t, err)
	require.NotNil(t, discordProvider.response.Data)
	assert.Equal(t, "Form 2", discordProvider.response.Data.Title.Val)

	row := (*discordProvider.response.Data.Components)[0].(*discord.ActionRowComponent)
	input := (*row)[0].(*discord.TextInputComponent)
	assert.Equal(t, discord.ComponentID("name_{{ 1 }}"), input.CustomID)
	assert.Equal(t, "Label 3", input.Label)
	assert.Equal(t, "Placeholder 4", input.Placeholder)
	assert.Equal(t, "Value 5", input.Value)
}

func TestFlowExecuteConditionCompareEquality(t *testing.T) {
	tests := []struct {
		name      string
		mode      ComparsionMode
		itemValue string
		expected  bool
	}{
		{name: "equal match", mode: ComparsionModeEqual, itemValue: "a", expected: true},
		{name: "equal mismatch", mode: ComparsionModeEqual, itemValue: "b", expected: false},
		{name: "not equal match", mode: ComparsionModeNotEqual, itemValue: "a", expected: false},
		{name: "not equal mismatch", mode: ComparsionModeNotEqual, itemValue: "b", expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			discordProvider := &TestDiscordProvider{}

			c := NewContext(
				ctx,
				5*time.Second,
				&TestContextData{},
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

			condition := &CompiledFlowNode{
				ID:   "1",
				Type: FlowNodeTypeControlConditionCompare,
				Data: FlowNodeData{ConditionBaseValue: "a"},
			}
			item := &CompiledFlowNode{
				ID:   "2",
				Type: FlowNodeTypeControlConditionItemCompare,
				Data: FlowNodeData{
					ConditionItemMode:  test.mode,
					ConditionItemValue: test.itemValue,
				},
				Parents: ConnectedFlowNodes{Default: []*CompiledFlowNode{condition}},
				Children: ConnectedFlowNodes{
					Default: []*CompiledFlowNode{{
						ID:   "3",
						Type: FlowNodeTypeActionResponseCreate,
						Data: FlowNodeData{
							MessageData: &message.MessageData{Content: "met"},
						},
					}},
				},
			}
			condition.Children.Default = []*CompiledFlowNode{item}

			node := CompiledFlowNode{
				ID:       "0",
				Type:     FlowNodeTypeEntryCommand,
				Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{condition}},
			}

			err := node.Execute(c)
			require.NoError(t, err)
			assert.Equal(t, test.expected, discordProvider.response.Data != nil)
		})
	}
}
