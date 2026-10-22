package flow

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type webhookContextData struct {
	TestContextData
	event *webhook.Event
}

func (d *webhookContextData) Event() ws.Event {
	return d.event
}

func executeWebhookResponse(t *testing.T, event *webhook.Event, data FlowNodeData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := NewContext(
		ctx,
		5*time.Second,
		&webhookContextData{event: event},
		FlowProviders{Log: &provider.MockLogProvider{}},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{"webhook": eval.NewWebhookEnv(event)}),
		nil,
	)
	defer c.Cancel()

	node := CompiledFlowNode{
		ID:   "0",
		Type: FlowNodeTypeEntryEvent,
		Data: FlowNodeData{EventType: EventTypeWebhook},
		Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{{
			ID:   "1",
			Type: FlowNodeTypeActionWebhookResponseCreate,
			Data: data,
		}}},
	}
	return node.Execute(c)
}

func TestWebhookResponseCreate(t *testing.T) {
	event := &webhook.Event{
		Body:      `{"code":"401"}`,
		Responder: webhook.NewResponder(),
	}

	err := executeWebhookResponse(t, event, FlowNodeData{
		WebhookResponseStatus:      "{{webhook.data.code}}",
		WebhookResponseContentType: "application/json",
		WebhookResponseBody:        `{"error":"{{webhook.data.code}}"}`,
	})
	require.NoError(t, err)

	resp := event.Responder.Wait(context.Background(), time.Second)
	require.NotNil(t, resp)
	assert.Equal(t, 401, resp.Status)
	assert.Equal(t, "application/json", resp.ContentType)
	assert.Equal(t, `{"error":"401"}`, resp.Body)
}

// Escapes like \n in JSON stay escapes, so the body stays valid JSON.
func TestWebhookResponseCreateKeepsEscapes(t *testing.T) {
	event := &webhook.Event{Body: `{"message":"a\nb"}`, Responder: webhook.NewResponder()}

	err := executeWebhookResponse(t, event, FlowNodeData{
		WebhookResponseStatus:      "200",
		WebhookResponseContentType: "application/json",
		WebhookResponseBody:        "{{webhook.body}}",
	})
	require.NoError(t, err)

	resp := event.Responder.Wait(context.Background(), time.Second)
	require.NotNil(t, resp)
	assert.Equal(t, `{"message":"a\nb"}`, resp.Body)
}

func TestWebhookResponseCreateDefaultsToText(t *testing.T) {
	event := &webhook.Event{Responder: webhook.NewResponder()}

	err := executeWebhookResponse(t, event, FlowNodeData{WebhookResponseStatus: "200"})
	require.NoError(t, err)

	resp := event.Responder.Wait(context.Background(), time.Second)
	require.NotNil(t, resp)
	assert.Equal(t, "text/plain", resp.ContentType)
	assert.Empty(t, resp.Body)
}

func TestWebhookResponseCreateErrors(t *testing.T) {
	cases := map[string]FlowNodeData{
		"empty status":    {},
		"invalid status":  {WebhookResponseStatus: "ok"},
		"redirect":        {WebhookResponseStatus: "302"},
		"html":            {WebhookResponseStatus: "200", WebhookResponseContentType: "text/html"},
		"body with 204":   {WebhookResponseStatus: "204", WebhookResponseBody: "a"},
		"invalid json":    {WebhookResponseStatus: "200", WebhookResponseContentType: "application/json", WebhookResponseBody: "{"},
		"status too high": {WebhookResponseStatus: "600"},
	}
	for name, data := range cases {
		event := &webhook.Event{Responder: webhook.NewResponder()}
		assert.Error(t, executeWebhookResponse(t, event, data), name)
	}
}

func TestWebhookResponseCreateAfterAnswer(t *testing.T) {
	event := &webhook.Event{Responder: webhook.NewResponder()}
	require.NoError(t, event.Responder.Respond(&webhook.Response{Status: 200, ContentType: "text/plain"}))

	// The flow continues, as when the request was answered depends on timing.
	err := executeWebhookResponse(t, event, FlowNodeData{WebhookResponseStatus: "200"})
	assert.NoError(t, err)

	// After a sleep the event has no responder.
	err = executeWebhookResponse(t, &webhook.Event{}, FlowNodeData{WebhookResponseStatus: "200"})
	assert.NoError(t, err)
}
