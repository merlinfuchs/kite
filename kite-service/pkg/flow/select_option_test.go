package flow

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

func selectMenuNode(handles ...string) *CompiledFlowNode {
	n := &CompiledFlowNode{
		Data: FlowNodeData{
			MessageData: &message.MessageData{
				Components: []message.ComponentData{
					{Type: message.ComponentTypeActionRow, Components: []message.ComponentData{
						{ID: 2, Type: message.ComponentTypeStringSelect, Options: []message.ComponentSelectOptionData{
							{ID: 3, Label: "Red", Value: "red"},
							{ID: 4, Label: "Blue", Value: "{{blue}}"},
						}},
					}},
				},
			},
		},
		Children: ConnectedFlowNodes{Handles: map[string][]*CompiledFlowNode{}},
	}
	for _, handle := range handles {
		n.Children.Handles[handle] = []*CompiledFlowNode{{ID: handle}}
	}
	return n
}

func selectInteraction(sent *discord.Message, values ...string) (*discord.InteractionEvent, discord.ComponentInteraction) {
	data := &discord.StringSelectInteraction{CustomID: "resume:abc_2", Values: values}
	return &discord.InteractionEvent{Data: data, Message: sent}, data
}

func TestSelectOptionRunsItsOwnOutput(t *testing.T) {
	n := selectMenuNode("component_2", "component_2_option_3")

	interaction, data := selectInteraction(nil, "red")
	if got := n.componentHandle(interaction, data, 2); got != "component_2_option_3" {
		t.Errorf("got %q, want the option's output", got)
	}
}

// Flows from before options had outputs compare interaction.value after the
// select menu's output, which keeps running for options that aren't connected.
func TestUnconnectedSelectOptionRunsMenuOutput(t *testing.T) {
	n := selectMenuNode("component_2", "component_2_option_4")

	interaction, data := selectInteraction(nil, "red")
	if got := n.componentHandle(interaction, data, 2); got != "component_2" {
		t.Errorf("got %q, want the select menu's output", got)
	}
}

// A value with a placeholder was filled in when sending, so it's found by its
// position in the sent menu.
func TestSelectOptionWithPlaceholderValueIsFoundInSentMenu(t *testing.T) {
	n := selectMenuNode("component_2_option_4")

	sent := &discord.Message{Components: discord.TopLevelComponents{
		&discord.ActionRowComponent{
			&discord.StringSelectComponent{CustomID: "resume:abc_2", Options: []discord.SelectOption{
				{Label: "Red", Value: "red"},
				{Label: "Blue", Value: "navy"},
			}},
		},
	}}

	interaction, data := selectInteraction(sent, "navy")
	if got := n.componentHandle(interaction, data, 2); got != "component_2_option_4" {
		t.Errorf("got %q, want the option's output", got)
	}
}

func TestButtonRunsComponentOutput(t *testing.T) {
	n := selectMenuNode("component_2_option_3")

	data := &discord.ButtonInteraction{CustomID: "resume:abc_2"}
	if got := n.componentHandle(&discord.InteractionEvent{Data: data}, data, 2); got != "component_2" {
		t.Errorf("got %q, want the component's output", got)
	}
}
