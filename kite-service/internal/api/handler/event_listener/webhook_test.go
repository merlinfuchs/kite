package eventlistener

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/core/engine"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type fakeListenerStore struct {
	store.EventListenerStore
	listeners []*model.EventListener
}

func (s *fakeListenerStore) EventListener(ctx context.Context, id string) (*model.EventListener, error) {
	for _, l := range s.listeners {
		if l != nil && l.ID == id {
			return l, nil
		}
	}
	return nil, store.ErrNotFound
}

// fakeRunner is the engine, which has the listeners of app "app" loaded with
// loadedSecret.
type fakeRunner struct {
	loadedSecret string
	err          error
	events       []*webhook.Event
}

func (r *fakeRunner) WebhookSecret(appID string, listenerID string) (string, bool) {
	if appID != "app" || (listenerID != "listener" && listenerID != "other-listener") {
		return "", false
	}
	return r.loadedSecret, true
}

func (r *fakeRunner) RunWebhook(ctx context.Context, appID string, listenerID string, event *webhook.Event) error {
	if r.err != nil {
		return r.err
	}
	r.events = append(r.events, event)
	return nil
}

func webhookTestListener() *model.EventListener {
	return &model.EventListener{
		ID:            "listener",
		AppID:         "app",
		Source:        model.EventSourceWebhook,
		Type:          model.EventListenerTypeWebhook,
		Enabled:       true,
		WebhookSecret: null.StringFrom("secret"),
	}
}

func webhookTestServer(listener *model.EventListener, runner *fakeRunner) http.Handler {
	return webhookTestMux(NewEventListenerHandler(&fakeListenerStore{listeners: []*model.EventListener{listener}}, runner))
}

func webhookTestMux(h *EventListenerHandler) http.Handler {
	mux := http.NewServeMux()
	group := handler.Group(mux, "/v1")
	group.Post("/apps/{appID}/webhooks/{listenerID}/{secret}", h.HandleEventListenerWebhook)
	return mux
}

func postWebhook(server http.Handler, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return rec
}

func TestWebhookRunsFlowWithRequest(t *testing.T) {
	runner := &fakeRunner{loadedSecret: "secret"}
	server := webhookTestServer(webhookTestListener(), runner)

	rec := postWebhook(server, "/v1/apps/app/webhooks/listener/secret?source=ci&tag=a&tag=b", `{"ok":true}`, map[string]string{
		"X-GitHub-Event": "push",
		"Cookie":         "kite-session=abc",
	})
	require.Equal(t, http.StatusAccepted, rec.Code, rec.Body.String())

	require.Len(t, runner.events, 1)
	event := runner.events[0]
	assert.Equal(t, `{"ok":true}`, event.Body)
	assert.Equal(t, "push", event.Headers["x-github-event"])
	assert.Equal(t, map[string]string{"source": "ci", "tag": "a,b"}, event.Query)
	assert.NotContains(t, event.Headers, "cookie", "cookies must not reach the flow")
}

func TestWebhookRejectsWrongURL(t *testing.T) {
	runner := &fakeRunner{loadedSecret: "secret"}
	server := webhookTestServer(webhookTestListener(), runner)

	for _, path := range []string{
		"/v1/apps/app/webhooks/listener/wrong",
		"/v1/apps/app/webhooks/other/secret",
		"/v1/apps/other/webhooks/listener/secret",
	} {
		rec := postWebhook(server, path, "", nil)
		assert.Equal(t, http.StatusNotFound, rec.Code, path)
	}
	assert.Empty(t, runner.events)
}

func TestWebhookChecksDatabaseAfterEngine(t *testing.T) {
	// The engine still has listeners loaded for a while after they change.
	cases := map[string]func(l *model.EventListener) *model.EventListener{
		"deleted":            func(l *model.EventListener) *model.EventListener { return nil },
		"disabled":           func(l *model.EventListener) *model.EventListener { l.Enabled = false; return l },
		"secret regenerated": func(l *model.EventListener) *model.EventListener { l.WebhookSecret = null.StringFrom("new"); return l },
		"moved to other app": func(l *model.EventListener) *model.EventListener { l.AppID = "other"; return l },
	}
	for name, change := range cases {
		runner := &fakeRunner{loadedSecret: "secret"}
		server := webhookTestServer(change(webhookTestListener()), runner)

		rec := postWebhook(server, "/v1/apps/app/webhooks/listener/secret", "", nil)
		assert.Equal(t, http.StatusNotFound, rec.Code, name)
		assert.Empty(t, runner.events, name)
	}
}

func TestWebhookRejectsLargeRequests(t *testing.T) {
	runner := &fakeRunner{loadedSecret: "secret"}
	server := webhookTestServer(webhookTestListener(), runner)

	rec := postWebhook(server, "/v1/apps/app/webhooks/listener/secret", strings.Repeat("a", webhook.MaxBodySize+1), nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	rec = postWebhook(server, "/v1/apps/app/webhooks/listener/secret", "", map[string]string{
		"X-Large": strings.Repeat("a", webhook.MaxMetadataSize),
	})
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	rec = postWebhook(server, "/v1/apps/app/webhooks/listener/secret", strings.Repeat("a", webhook.MaxBodySize), nil)
	assert.Equal(t, http.StatusAccepted, rec.Code, "a body at the limit is accepted")

	assert.Len(t, runner.events, 1)
}

func TestWebhookRateLimitIsPerApp(t *testing.T) {
	other := webhookTestListener()
	other.ID = "other-listener"

	runner := &fakeRunner{loadedSecret: "secret"}
	server := webhookTestMux(NewEventListenerHandler(
		&fakeListenerStore{listeners: []*model.EventListener{webhookTestListener(), other}},
		runner,
	))

	// Requests with a wrong secret don't count against the app.
	for i := 0; i < webhookRateLimit; i++ {
		postWebhook(server, "/v1/apps/app/webhooks/listener/wrong", "", nil)
	}

	// Both listeners of the app share one limit.
	for i := 0; i < webhookRateLimit; i++ {
		path := "/v1/apps/app/webhooks/listener/secret"
		if i%2 == 1 {
			path = "/v1/apps/app/webhooks/other-listener/secret"
		}
		rec := postWebhook(server, path, "", nil)
		require.Equal(t, http.StatusAccepted, rec.Code, "request %d", i)
	}

	for _, path := range []string{
		"/v1/apps/app/webhooks/listener/secret",
		"/v1/apps/app/webhooks/other-listener/secret",
	} {
		rec := postWebhook(server, path, "", nil)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code, path)
		assert.NotEmpty(t, rec.Header().Get("Retry-After"))
	}
	assert.Len(t, runner.events, webhookRateLimit)
}

func TestWebhookAppOffline(t *testing.T) {
	runner := &fakeRunner{loadedSecret: "secret", err: engine.ErrWebhookAppOffline}
	server := webhookTestServer(webhookTestListener(), runner)

	rec := postWebhook(server, "/v1/apps/app/webhooks/listener/secret", "", nil)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestNewWebhookSecret(t *testing.T) {
	assert.False(t, newWebhookSecret(model.EventSourceDiscord).Valid)
	assert.False(t, newWebhookSecret(model.EventSourceSchedule).Valid)

	a := newWebhookSecret(model.EventSourceWebhook)
	b := newWebhookSecret(model.EventSourceWebhook)
	require.True(t, a.Valid)
	assert.NotEqual(t, a.String, b.String)
	// The secret is a path segment of the webhook URL.
	assert.NotContains(t, a.String, "/")
	assert.GreaterOrEqual(t, len(a.String), 32)
}
