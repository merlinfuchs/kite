package marketplace

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func commandFlow(name string, action flow.FlowNodeType) flow.FlowData {
	return flow.FlowData{
		Nodes: []flow.FlowNode{
			{
				ID:   "0",
				Type: flow.FlowNodeTypeEntryCommand,
				Data: flow.FlowNodeData{Name: name, Description: "Does " + name},
			},
			{
				ID:   "1",
				Type: action,
				Data: flow.FlowNodeData{MessageData: &message.MessageData{Content: "Hi"}},
			},
		},
		Edges: []flow.FlowEdge{{Source: "0", Target: "1"}},
	}
}

func scheduleFlow() flow.FlowData {
	return flow.FlowData{
		Nodes: []flow.FlowNode{
			{
				ID:   "0",
				Type: flow.FlowNodeTypeEntryEvent,
				Data: flow.FlowNodeData{
					EventType:         flow.EventTypeScheduleCron,
					EventScheduleCron: "*/5 * * * *",
					Description:       "Every five minutes",
				},
			},
			{ID: "1", Type: flow.FlowNodeTypeActionLog},
		},
		Edges: []flow.FlowEdge{{Source: "0", Target: "1"}},
	}
}

func TestCompileListingItemsModule(t *testing.T) {
	content, err := compileListingItems([]wire.MarketplaceListingItemRequest{
		{Type: "command", FlowSource: commandFlow("ping", flow.FlowNodeTypeActionResponseCreate)},
		{Type: "event_listener", Source: "schedule", FlowSource: scheduleFlow()},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, content.commandCount)
	assert.Equal(t, 1, content.eventListenerCount)
	assert.Equal(t, "ping", content.items[0].Name)
	assert.Equal(t, "Does ping", content.items[0].Description)
	assert.Equal(t, "schedule", content.items[1].Source)
	assert.Equal(t, []string{
		string(flow.FlowNodeTypeActionLog),
		string(flow.FlowNodeTypeActionResponseCreate),
		string(flow.FlowNodeTypeEntryCommand),
		string(flow.FlowNodeTypeEntryEvent),
	}, content.blockTypes)

	listing := &model.MarketplaceListing{CommandCount: 1, EventListenerCount: 1}
	assert.True(t, listing.IsModule())
}

func TestCompileListingItemsDuplicateCommand(t *testing.T) {
	_, err := compileListingItems([]wire.MarketplaceListingItemRequest{
		{Type: "command", FlowSource: commandFlow("ping", flow.FlowNodeTypeActionResponseCreate)},
		{Type: "command", FlowSource: commandFlow("ping", flow.FlowNodeTypeActionResponseCreate)},
	})
	assert.Error(t, err)
}

func TestCompileListingItemsWrongType(t *testing.T) {
	// A command flow published as an event listener has no event entry.
	_, err := compileListingItems([]wire.MarketplaceListingItemRequest{
		{Type: "event_listener", Source: "discord", FlowSource: commandFlow("ping", flow.FlowNodeTypeActionResponseCreate)},
	})
	assert.Error(t, err)
}

func TestCompileListingItemsWrongSource(t *testing.T) {
	_, err := compileListingItems([]wire.MarketplaceListingItemRequest{
		{Type: "event_listener", Source: "discord", FlowSource: scheduleFlow()},
	})
	assert.Error(t, err)
}

func messageItem(sourceID string) wire.MarketplaceListingItemRequest {
	return wire.MarketplaceListingItemRequest{
		Type:     "message",
		Name:     "Welcome",
		SourceID: sourceID,
		MessageData: &message.MessageData{
			Content: "Welcome!",
			Components: []message.ComponentData{{
				Type: 1,
				Components: []message.ComponentData{
					{Type: 2, Label: "Rules", FlowSourceID: "button1"},
				},
			}},
		},
		MessageFlowSources: map[string]flow.FlowData{
			"button1": {
				Nodes: []flow.FlowNode{
					{ID: "0", Type: flow.FlowNodeTypeEntryComponentButton},
					{ID: "1", Type: flow.FlowNodeTypeActionLog},
				},
				Edges: []flow.FlowEdge{{Source: "0", Target: "1"}},
			},
			// Not used by any component, dropped like in a regular import.
			"unused": {},
		},
	}
}

func TestCompileListingItemsMessageModule(t *testing.T) {
	content, err := compileListingItems([]wire.MarketplaceListingItemRequest{
		{Type: "command", FlowSource: commandFlow("welcome", flow.FlowNodeTypeActionResponseCreate)},
		messageItem("msg1"),
	})
	require.NoError(t, err)

	assert.Equal(t, 1, content.commandCount)
	assert.Equal(t, 1, content.messageCount)

	msg := content.items[1]
	assert.Equal(t, model.MarketplaceListingItemTypeMessage, msg.Type)
	assert.Equal(t, "Welcome", msg.Name)
	assert.Equal(t, "msg1", msg.SourceID)
	require.NotNil(t, msg.MessageData)
	assert.Contains(t, msg.MessageFlowSources, "button1")
	assert.NotContains(t, msg.MessageFlowSources, "unused")
	assert.Contains(t, content.blockTypes, string(flow.FlowNodeTypeEntryComponentButton))

	listing := &model.MarketplaceListing{CommandCount: 1, MessageCount: 1}
	assert.True(t, listing.IsModule())
	assert.Equal(t, "module", wire.MarketplaceListingToWire(listing, nil, false).Kind)

	single := &model.MarketplaceListing{MessageCount: 1}
	assert.Equal(t, "message", wire.MarketplaceListingToWire(single, nil, false).Kind)
}

func TestCompileListingItemsDuplicateMessage(t *testing.T) {
	_, err := compileListingItems([]wire.MarketplaceListingItemRequest{
		messageItem("msg1"),
		messageItem("msg1"),
	})
	assert.Error(t, err)
}

func TestMarketplaceListingItemRequestMessageNeedsData(t *testing.T) {
	item := messageItem("msg1")
	item.MessageData = nil
	assert.Error(t, item.Validate())

	item = messageItem("msg1")
	item.Name = ""
	assert.Error(t, item.Validate())

	assert.NoError(t, messageItem("msg1").Validate())
}
