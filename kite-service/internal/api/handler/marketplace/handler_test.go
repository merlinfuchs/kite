package marketplace

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
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

func httpCommandFlow(name string, httpData flow.HTTPRequestData) flow.FlowData {
	return flow.FlowData{
		Nodes: []flow.FlowNode{
			{
				ID:   "0",
				Type: flow.FlowNodeTypeEntryCommand,
				Data: flow.FlowNodeData{Name: name, Description: "Does " + name},
			},
			{
				ID:   "1",
				Type: flow.FlowNodeTypeActionHTTPRequest,
				Data: flow.FlowNodeData{HTTPRequestData: &httpData},
			},
		},
		Edges: []flow.FlowEdge{{Source: "0", Target: "1"}},
	}
}

func TestCompileListingItemsSanitizesCredentials(t *testing.T) {
	content, err := compileListingItems([]wire.MarketplaceListingItemRequest{
		{
			Type: "command",
			FlowSource: httpCommandFlow("test_http", flow.HTTPRequestData{
				URL:    "https://discord.com/api/webhooks/1234567890/abc123secretToken",
				Method: "POST",
				Headers: []flow.HTTPRequestDataKeyValue{
					{Key: "Authorization", Value: "Bearer raw_secret_123"},
					{Key: "X-Api-Key", Value: "super_secret_api_key"},
					{Key: "Content-Type", Value: "application/json"},
				},
				Query: []flow.HTTPRequestDataKeyValue{
					{Key: "api_key", Value: "query_secret_key"},
					{Key: "page", Value: "1"},
				},
			}),
		},
	})
	require.NoError(t, err)
	require.Len(t, content.items, 1)

	node := content.items[0].FlowSource.Nodes[1]
	require.NotNil(t, node.Data.HTTPRequestData)

	// Webhook token must be redacted
	assert.Equal(t, "https://discord.com/api/webhooks/1234567890/[REDACTED]", node.Data.HTTPRequestData.URL)

	// Hardcoded Authorization and X-Api-Key headers must be stripped, Content-Type kept
	assert.Len(t, node.Data.HTTPRequestData.Headers, 1)
	assert.Equal(t, "Content-Type", node.Data.HTTPRequestData.Headers[0].Key)
	assert.Equal(t, "application/json", node.Data.HTTPRequestData.Headers[0].Value)

	// Sensitive query param api_key must be redacted, page kept
	assert.Len(t, node.Data.HTTPRequestData.Query, 2)
	assert.Equal(t, "api_key", node.Data.HTTPRequestData.Query[0].Key)
	assert.Equal(t, "[REDACTED]", node.Data.HTTPRequestData.Query[0].Value)
	assert.Equal(t, "page", node.Data.HTTPRequestData.Query[1].Key)
	assert.Equal(t, "1", node.Data.HTTPRequestData.Query[1].Value)
}

func TestCompileListingItemsPreservesSecrets(t *testing.T) {
	content, err := compileListingItems([]wire.MarketplaceListingItemRequest{
		{
			Type: "command",
			FlowSource: httpCommandFlow("test_secrets", flow.HTTPRequestData{
				URL:    "https://example.com/api?token={{secrets.API_TOKEN}}",
				Method: "GET",
				Headers: []flow.HTTPRequestDataKeyValue{
					{Key: "Authorization", Value: "Bearer {{secrets.AUTH_TOKEN}}"},
					{Key: "X-Api-Key", Value: "{{ secrets.CUSTOM_KEY }}"},
				},
			}),
		},
	})
	require.NoError(t, err)
	require.Len(t, content.items, 1)

	node := content.items[0].FlowSource.Nodes[1]
	require.NotNil(t, node.Data.HTTPRequestData)

	// Headers referencing {{secrets...}} must NOT be stripped
	assert.Len(t, node.Data.HTTPRequestData.Headers, 2)
	assert.Equal(t, "Authorization", node.Data.HTTPRequestData.Headers[0].Key)
	assert.Equal(t, "Bearer {{secrets.AUTH_TOKEN}}", node.Data.HTTPRequestData.Headers[0].Value)
	assert.Equal(t, "X-Api-Key", node.Data.HTTPRequestData.Headers[1].Key)
	assert.Equal(t, "{{ secrets.CUSTOM_KEY }}", node.Data.HTTPRequestData.Headers[1].Value)

	// URL referencing secret must be preserved
	assert.Equal(t, "https://example.com/api?token={{secrets.API_TOKEN}}", node.Data.HTTPRequestData.URL)
}

type mockMarketplaceStore struct {
	store.MarketplaceStore
	reports []*model.MarketplaceReport
	listing *model.MarketplaceListingWithAuthor
}

func (m *mockMarketplaceStore) MarketplaceListing(ctx context.Context, id string) (*model.MarketplaceListingWithAuthor, error) {
	if m.listing != nil && m.listing.ID == id {
		return m.listing, nil
	}
	return nil, store.ErrNotFound
}

func (m *mockMarketplaceStore) CreateMarketplaceReport(ctx context.Context, report *model.MarketplaceReport) (*model.MarketplaceReport, error) {
	for _, r := range m.reports {
		if r.ListingID == report.ListingID && r.ReporterUserID == report.ReporterUserID && !r.ResolvedAt.Valid {
			return nil, store.ErrAlreadyExists
		}
	}
	m.reports = append(m.reports, report)
	return report, nil
}

func (m *mockMarketplaceStore) CountOpenMarketplaceReportsByListing(ctx context.Context, listingID string) (int, error) {
	count := 0
	for _, r := range m.reports {
		if r.ListingID == listingID && !r.ResolvedAt.Valid {
			count++
		}
	}
	return count, nil
}

func TestHandleMarketplaceReportDuplicate(t *testing.T) {
	mockStore := &mockMarketplaceStore{
		listing: &model.MarketplaceListingWithAuthor{
			MarketplaceListing: model.MarketplaceListing{
				ID:           "listing_1",
				AuthorUserID: "author_1",
				Status:       model.MarketplaceListingStatusApproved,
			},
		},
	}

	h := NewMarketplaceHandler(
		MarketplaceHandlerConfig{},
		mockStore,
		nil,
	)

	mux := http.NewServeMux()
	mux.Handle("POST /api/marketplace/listings/{listingID}/reports", handler.APIHandler(func(c *handler.Context) error {
		c.Session = &model.Session{UserID: "reporter_user_1"}
		return handler.TypedWithBody(h.HandleMarketplaceReportCreate)(c)
	}))

	// First report succeeds
	req1 := httptest.NewRequest("POST", "/api/marketplace/listings/listing_1/reports", strings.NewReader(`{"reason":"malicious code"}`))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	mux.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	// Second report while first is unresolved fails with 400 already_reported
	req2 := httptest.NewRequest("POST", "/api/marketplace/listings/listing_1/reports", strings.NewReader(`{"reason":"another report"}`))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusBadRequest, rec2.Code)

	var errRes map[string]any
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &errRes))
	errObj, ok := errRes["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "already_reported", errObj["code"])
}
