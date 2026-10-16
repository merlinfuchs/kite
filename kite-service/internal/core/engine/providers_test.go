package engine

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

type fakeMessageStore struct {
	store.MessageStore
	messages map[string]*model.Message
}

func (f *fakeMessageStore) Message(ctx context.Context, appID string, id string) (*model.Message, error) {
	msg, ok := f.messages[id]
	if !ok || msg.AppID != appID {
		return nil, store.ErrNotFound
	}
	return msg, nil
}

// Template IDs in flow data are user-authored, so a flow must not be able to
// send or link another app's template.
func TestMessageTemplateProviderScopedToApp(t *testing.T) {
	messages := &fakeMessageStore{messages: map[string]*model.Message{
		"own":   {ID: "own", AppID: "app"},
		"other": {ID: "other", AppID: "other_app"},
	}}
	p := NewMessageTemplateProvider("app", messages, nil)

	if _, err := p.MessageTemplate(context.Background(), "own"); err != nil {
		t.Errorf("own template: %v", err)
	}
	if _, err := p.MessageTemplate(context.Background(), "other"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("other app's template: got %v, want ErrNotFound", err)
	}
}

// The Discord API Request block relies on the session's client to add the
// bot's token, and on failed requests turning into errors.
func TestDiscordProviderAPIRequest(t *testing.T) {
	var got *http.Request
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		gotBody, _ = io.ReadAll(r.Body)
		if r.URL.Path == "/api/v9/channels/2" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code":10003,"message":"Unknown Channel"}`))
			return
		}
		w.Write([]byte(`{"id":"3"}`))
	}))
	defer server.Close()

	endpoint := api.Endpoint
	api.Endpoint = server.URL + api.Path + "/"
	defer func() { api.Endpoint = endpoint }()

	p := NewDiscordProvider("app", nil, nil, nil, nil, state.New("Bot token"))

	body, err := p.APIRequest(context.Background(), provider.DiscordAPIRequest{
		Method: http.MethodPost,
		Path:   "/channels/1/messages?limit=5",
		Body:   []byte(`{"content":"hi"}`),
		Reason: "because",
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if string(body) != `{"id":"3"}` {
		t.Errorf("response: got %s", body)
	}
	if got.Method != http.MethodPost || got.URL.Path != "/api/v9/channels/1/messages" || got.URL.RawQuery != "limit=5" {
		t.Errorf("request: got %s %s", got.Method, got.URL)
	}
	if got.Header.Get("Authorization") != "Bot token" {
		t.Errorf("authorization: got %q", got.Header.Get("Authorization"))
	}
	if got.Header.Get("X-Audit-Log-Reason") != "because" {
		t.Errorf("audit log reason: got %q", got.Header.Get("X-Audit-Log-Reason"))
	}
	if got.Header.Get("Content-Type") != "application/json" || string(gotBody) != `{"content":"hi"}` {
		t.Errorf("body: got %q %s", got.Header.Get("Content-Type"), gotBody)
	}

	_, err = p.APIRequest(context.Background(), provider.DiscordAPIRequest{
		Method: http.MethodGet,
		Path:   "/channels/2",
	})
	if err == nil || !strings.Contains(err.Error(), "Discord 404 error: Unknown Channel") {
		t.Errorf("failed request: got %v", err)
	}
}

// The gateway cache has no member counts, so the stats have to ask the API
// for them.
func TestDiscordProviderBotStats(t *testing.T) {
	var got *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Write([]byte(`[
			{"id":"1","approximate_member_count":10},
			{"id":"2","approximate_member_count":32}
		]`))
	}))
	defer server.Close()

	endpoint := api.EndpointMe
	api.EndpointMe = server.URL + api.Path + "/users/@me"
	defer func() { api.EndpointMe = endpoint }()

	connections := NewConnectionTracker()
	connections.Connected("app")

	p := NewDiscordProvider("app", nil, nil, nil, connections, state.New("Bot token"))

	stats, err := p.BotStats(context.Background())
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if got.URL.Path != "/api/v9/users/@me/guilds" || got.URL.Query().Get("with_counts") != "true" {
		t.Errorf("request: got %s", got.URL)
	}
	if stats.GuildCount != 2 || stats.MemberCount != 42 {
		t.Errorf("counts: got %d servers and %d members", stats.GuildCount, stats.MemberCount)
	}
	if stats.ConnectedAt != connections.ConnectedAt("app") {
		t.Errorf("connected at: got %s", stats.ConnectedAt)
	}
	// The session never connected, so there's no heartbeat to measure.
	if stats.Latency != 0 {
		t.Errorf("latency: got %s", stats.Latency)
	}
}
