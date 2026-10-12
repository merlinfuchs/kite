package flow

import (
	"context"
	"encoding/json"
	"strings"
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

func (d *TestContextData) MessageID() discord.MessageID {
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

func executeModal(t *testing.T, modal *ModalData) (*TestDiscordProvider, error) {
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
					Data: FlowNodeData{ModalData: modal},
				},
			},
		},
	}

	return discordProvider, node.Execute(c)
}

// modalComponentsJSON returns the components of the modal response as Discord
// receives them.
func modalComponentsJSON(t *testing.T, discordProvider *TestDiscordProvider) []map[string]any {
	require.NotNil(t, discordProvider.response.Data)
	require.NotNil(t, discordProvider.response.Data.Components)

	b, err := json.Marshal(discordProvider.response.Data.Components)
	require.NoError(t, err)

	var res []map[string]any
	require.NoError(t, json.Unmarshal(b, &res))
	return res
}

func TestFlowExecuteModalEvaluatesTemplates(t *testing.T) {
	discordProvider, err := executeModal(t, &ModalData{
		Title: "Form {{ 1 + 1 }}",
		Components: []ModalComponentData{{
			Components: []ModalComponentData{{
				CustomID:    "name_{{ 1 }}",
				Style:       1,
				Label:       "Label {{ 2 + 1 }}",
				Placeholder: "Placeholder {{ 4 }}",
				Value:       "Value {{ 5 }}",
			}, {
				CustomID: "empty",
				Style:    1,
				Label:    "Empty",
			}},
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, "Form 2", discordProvider.response.Data.Title.Val)

	components := modalComponentsJSON(t, discordProvider)
	require.Len(t, components, 2)

	// A legacy row becomes one label per text input.
	assert.EqualValues(t, discord.LabelComponentType, components[0]["type"])
	assert.Equal(t, "Label 3", components[0]["label"])

	input := components[0]["component"].(map[string]any)
	assert.EqualValues(t, discord.TextInputComponentType, input["type"])
	assert.Equal(t, "name_{{ 1 }}", input["custom_id"])
	assert.Equal(t, "Placeholder 4", input["placeholder"])
	assert.Equal(t, "Value 5", input["value"])
	assert.NotContains(t, input, "label")

	assert.Equal(t, "Empty", components[1]["label"])
	empty := components[1]["component"].(map[string]any)
	assert.NotContains(t, empty, "placeholder")
	assert.NotContains(t, empty, "value")
}

func TestFlowExecuteModalComponents(t *testing.T) {
	discordProvider, err := executeModal(t, &ModalData{
		Title: "Form",
		Components: []ModalComponentData{
			{
				Type:    ModalComponentTypeTextDisplay,
				Content: "Hello {{ 1 + 1 }}",
			},
			{
				Type:        ModalComponentTypeLabel,
				Label:       "Color",
				Description: "Pick {{ 1 }}",
				Components: []ModalComponentData{{
					Type:      ModalComponentTypeStringSelect,
					CustomID:  "color",
					MaxValues: 2,
					Options: []ModalComponentOptionData{
						{Label: "Red", Value: "red"},
						{Label: "Blue", Default: true},
					},
				}},
			},
			{
				Type:  ModalComponentTypeLabel,
				Label: "Channel",
				Components: []ModalComponentData{{
					Type:         ModalComponentTypeChannelSelect,
					CustomID:     "channel",
					Required:     true,
					ChannelTypes: []int{0},
				}},
			},
			{
				Type:  ModalComponentTypeLabel,
				Label: "Agree",
				Components: []ModalComponentData{{
					Type:     ModalComponentTypeCheckbox,
					CustomID: "agree",
					Default:  true,
				}},
			},
		},
	})
	require.NoError(t, err)

	components := modalComponentsJSON(t, discordProvider)
	require.Len(t, components, 4)

	assert.EqualValues(t, discord.TextDisplayComponentType, components[0]["type"])
	assert.Equal(t, "Hello 2", components[0]["content"])

	assert.Equal(t, "Pick 1", components[1]["description"])
	sel := components[1]["component"].(map[string]any)
	assert.EqualValues(t, discord.StringSelectComponentType, sel["type"])
	// An optional select has to send required: false and min_values: 0, as
	// Discord defaults them to true and 1.
	assert.Equal(t, false, sel["required"])
	assert.EqualValues(t, 0, sel["min_values"])
	assert.EqualValues(t, 2, sel["max_values"])
	options := sel["options"].([]any)
	assert.Equal(t, "Blue", options[1].(map[string]any)["value"])
	assert.Equal(t, true, options[1].(map[string]any)["default"])

	channel := components[2]["component"].(map[string]any)
	assert.EqualValues(t, discord.ChannelSelectComponentType, channel["type"])
	assert.Equal(t, true, channel["required"])
	assert.NotContains(t, channel, "min_values")
	assert.Equal(t, []any{float64(0)}, channel["channel_types"])

	checkbox := components[3]["component"].(map[string]any)
	assert.EqualValues(t, discord.CheckboxComponentType, checkbox["type"])
	assert.Equal(t, true, checkbox["default"])
}

func modalWithInput(input ModalComponentData) *ModalData {
	return &ModalData{Title: "Form", Components: []ModalComponentData{
		{Type: ModalComponentTypeLabel, Label: "Input", Components: []ModalComponentData{input}},
	}}
}

func TestFlowExecuteModalLimits(t *testing.T) {
	// Discord lets every option of a checkbox group be picked by default.
	_, err := executeModal(t, modalWithInput(ModalComponentData{Type: ModalComponentTypeCheckboxGroup, CustomID: "name", Options: []ModalComponentOptionData{
		{Label: "A", Default: true},
		{Label: "B", Default: true},
	}}))
	require.NoError(t, err)

	_, err = executeModal(t, modalWithInput(ModalComponentData{Type: ModalComponentTypeStringSelect, CustomID: "name", Placeholder: strings.Repeat("a", 150), MaxValues: 2, Options: []ModalComponentOptionData{
		{Label: "A", Default: true},
		{Label: "B", Default: true},
	}}))
	require.NoError(t, err)

	_, err = executeModal(t, modalWithInput(ModalComponentData{Type: ModalComponentTypeTextInput, CustomID: "name", Placeholder: strings.Repeat("a", 100)}))
	require.NoError(t, err)
}

func TestFlowExecuteModalInvalid(t *testing.T) {
	tests := []struct {
		name  string
		modal *ModalData
	}{
		{
			name:  "no components",
			modal: &ModalData{Title: "Form"},
		},
		{
			name: "only text",
			modal: &ModalData{Title: "Form", Components: []ModalComponentData{
				{Type: ModalComponentTypeTextDisplay, Content: "Hi"},
			}},
		},
		{
			name: "label without input",
			modal: &ModalData{Title: "Form", Components: []ModalComponentData{
				{Type: ModalComponentTypeLabel, Label: "Name"},
			}},
		},
		{
			name: "select without options",
			modal: &ModalData{Title: "Form", Components: []ModalComponentData{
				{Type: ModalComponentTypeLabel, Label: "Name", Components: []ModalComponentData{
					{Type: ModalComponentTypeStringSelect, CustomID: "name"},
				}},
			}},
		},
		{
			name: "radio group with one option",
			modal: &ModalData{Title: "Form", Components: []ModalComponentData{
				{Type: ModalComponentTypeLabel, Label: "Name", Components: []ModalComponentData{
					{Type: ModalComponentTypeRadioGroup, CustomID: "name", Options: []ModalComponentOptionData{{Label: "A"}}},
				}},
			}},
		},
		{
			name: "min above max",
			modal: &ModalData{Title: "Form", Components: []ModalComponentData{
				{Type: ModalComponentTypeLabel, Label: "Name", Components: []ModalComponentData{
					{Type: ModalComponentTypeUserSelect, CustomID: "name", MinValues: 3, MaxValues: 2},
				}},
			}},
		},
		{
			name:  "min above default max",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeUserSelect, CustomID: "name", MinValues: 3}),
		},
		{
			name: "duplicate option values",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeStringSelect, CustomID: "name", Options: []ModalComponentOptionData{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "{{ 'a' }}"},
			}}),
		},
		{
			name: "option value same as label of another",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeCheckboxGroup, CustomID: "name", Options: []ModalComponentOptionData{
				{Label: "A"},
				{Label: "B", Value: "A"},
			}}),
		},
		{
			name: "duplicate identifiers",
			modal: &ModalData{Title: "Form", Components: []ModalComponentData{
				{Type: ModalComponentTypeLabel, Label: "A", Components: []ModalComponentData{
					{Type: ModalComponentTypeTextInput, CustomID: "name"},
				}},
				{Type: ModalComponentTypeLabel, Label: "B", Components: []ModalComponentData{
					{Type: ModalComponentTypeCheckbox, CustomID: "name"},
				}},
			}},
		},
		{
			name: "max values above options",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeStringSelect, CustomID: "name", MaxValues: 3, Options: []ModalComponentOptionData{
				{Label: "A"},
				{Label: "B"},
			}}),
		},
		{
			name: "min values above options",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeCheckboxGroup, CustomID: "name", MinValues: 2, Options: []ModalComponentOptionData{
				{Label: "A"},
			}}),
		},
		{
			name: "radio group with two defaults",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeRadioGroup, CustomID: "name", Options: []ModalComponentOptionData{
				{Label: "A", Default: true},
				{Label: "B", Default: true},
			}}),
		},
		{
			name: "single select with two defaults",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeStringSelect, CustomID: "name", Options: []ModalComponentOptionData{
				{Label: "A", Default: true},
				{Label: "B", Default: true},
			}}),
		},
		{
			name: "checkbox group with more defaults than max",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeCheckboxGroup, CustomID: "name", MaxValues: 1, Options: []ModalComponentOptionData{
				{Label: "A", Default: true},
				{Label: "B", Default: true},
			}}),
		},
		{
			name:  "min length above max length",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeTextInput, CustomID: "name", MinLength: 10, MaxLength: 5}),
		},
		{
			name:  "text input placeholder too long",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeTextInput, CustomID: "name", Placeholder: strings.Repeat("a", 101)}),
		},
		{
			name:  "select placeholder too long",
			modal: modalWithInput(ModalComponentData{Type: ModalComponentTypeRoleSelect, CustomID: "name", Placeholder: strings.Repeat("a", 151)}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			discordProvider, err := executeModal(t, test.modal)
			require.Error(t, err)
			assert.NotEqual(t, api.ModalResponse, discordProvider.response.Type)
		})
	}
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

// Every block is defined in block_definitions.json. Custom blocks need a
// handler, except options, which configure the entry block and never run.
func TestEveryBlockRuns(t *testing.T) {
	for _, nodeType := range flowNodeTypeConstants(t) {
		assert.Containsf(t, blockDefinitions, FlowNodeType(nodeType), "%s has no definition", nodeType)
	}

	for nodeType, block := range blockDefinitions {
		_, handled := nodeHandlers[nodeType]
		switch {
		case block.Run.Kind == "request":
			assert.Falsef(t, handled, "%s is a request but has a handler", nodeType)
		case strings.HasPrefix(string(nodeType), "option_"):
			assert.Falsef(t, handled, "%s is an option but has a handler", nodeType)
		default:
			assert.Truef(t, handled, "%s has no handler", nodeType)
		}
	}
	for nodeType := range nodeHandlers {
		assert.Containsf(t, blockDefinitions, nodeType, "%s has no definition", nodeType)
	}
}

// Custom blocks compute their credits in CreditsCost, which has to match
// what the editor shows.
func TestBlockCredits(t *testing.T) {
	for nodeType, block := range blockDefinitions {
		if block.Credits == nil {
			continue
		}
		node := &CompiledFlowNode{Type: nodeType}
		assert.Equalf(t, *block.Credits, node.CreditsCost(), "%s", nodeType)
	}
}
