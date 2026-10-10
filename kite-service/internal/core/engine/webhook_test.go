package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
	"gopkg.in/guregu/null.v4"
)

func webhookListener() *model.EventListener {
	return &model.EventListener{
		ID:            "listener",
		AppID:         "app",
		Source:        model.EventSourceWebhook,
		Type:          model.EventListenerTypeWebhook,
		Enabled:       true,
		WebhookSecret: null.StringFrom("secret"),
		FlowSource: flow.FlowData{
			Nodes: []flow.FlowNode{
				{ID: "entry", Type: flow.FlowNodeTypeEntryEvent, Data: flow.FlowNodeData{
					EventType:   flow.EventTypeWebhook,
					Description: "test",
				}},
				{ID: "log", Type: flow.FlowNodeTypeActionLog, Data: flow.FlowNodeData{
					LogLevel:   provider.LogLevelInfo,
					LogMessage: "{{webhook.headers['x-event']}} to {{webhook.data.repository}}",
				}},
			},
			Edges: []flow.FlowEdge{{Source: "entry", Target: "log"}},
		},
	}
}

func TestWebhookRunnerRunsFlow(t *testing.T) {
	listeners := &fakeScheduleListenerStore{
		lastRuns:  make(map[string]time.Time),
		listeners: []*model.EventListener{webhookListener()},
	}
	logs := &fakeLogStore{entries: make(chan model.LogEntry, 1)}

	e := newScheduleTestEngine(listeners, logs)
	e.populate(context.Background())

	r := e.WebhookRunner(fakeSessions{})

	secret, ok := r.WebhookSecret("app", "listener")
	if !ok || secret != "secret" {
		t.Fatalf("WebhookSecret = %q, %v", secret, ok)
	}

	err := r.RunWebhook(context.Background(), "app", "listener", &webhook.Event{
		Headers: map[string]string{"x-event": "push"},
		Body:    `{"repository":"kite"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case entry := <-logs.entries:
		if entry.Message != "push to kite" {
			t.Fatalf("log message = %q", entry.Message)
		}
		if entry.EventListenerID.String != "listener" {
			t.Fatalf("log not attributed to the listener: %+v", entry)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("webhook flow didn't run")
	}
}

func TestWebhookRunnerOnlyRunsWebhookListeners(t *testing.T) {
	listeners := &fakeScheduleListenerStore{
		lastRuns:  make(map[string]time.Time),
		listeners: []*model.EventListener{scheduledListener("* * * * *")},
	}
	logs := &fakeLogStore{entries: make(chan model.LogEntry, 1)}

	e := newScheduleTestEngine(listeners, logs)
	e.populate(context.Background())

	r := e.WebhookRunner(fakeSessions{})

	if _, ok := r.WebhookSecret("app", "listener"); ok {
		t.Fatal("a scheduled listener has a webhook secret")
	}
	if _, ok := r.WebhookSecret("other", "listener"); ok {
		t.Fatal("found the listener under another app")
	}

	err := r.RunWebhook(context.Background(), "app", "listener", &webhook.Event{})
	if !errors.Is(err, ErrWebhookListenerNotFound) {
		t.Fatalf("err = %v, want ErrWebhookListenerNotFound", err)
	}
}

func TestWebhookRunnerNeedsSession(t *testing.T) {
	listeners := &fakeScheduleListenerStore{
		lastRuns:  make(map[string]time.Time),
		listeners: []*model.EventListener{webhookListener()},
	}
	logs := &fakeLogStore{entries: make(chan model.LogEntry, 1)}

	e := newScheduleTestEngine(listeners, logs)
	e.populate(context.Background())

	r := e.WebhookRunner(&flakySessions{})

	err := r.RunWebhook(context.Background(), "app", "listener", &webhook.Event{})
	if !errors.Is(err, ErrWebhookAppOffline) {
		t.Fatalf("err = %v, want ErrWebhookAppOffline", err)
	}
}
