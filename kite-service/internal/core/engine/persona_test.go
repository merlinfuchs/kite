package engine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type staticFeatureProvider struct {
	features model.Features
}

func (p staticFeatureProvider) AppFeatures(ctx context.Context, appID string) (model.Features, error) {
	return p.features, nil
}

func (p staticFeatureProvider) AppFeaturesForApps(ctx context.Context, appIDs []string) (map[string]model.Features, error) {
	return nil, nil
}

// fakeWebhookDiscord serves the requests of a persona message. Channel 1 is a
// text channel with a webhook of another bot, channel 3 a thread in it.
type fakeWebhookDiscord struct {
	mu         sync.Mutex
	requests   []string
	executions []*http.Request
	bodies     []map[string]any
	// deleted makes the next execution fail like a deleted webhook.
	deleted bool
	created int
}

func (f *fakeWebhookDiscord) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	path := strings.TrimPrefix(r.URL.Path, "/api/v9")
	f.requests = append(f.requests, r.Method+" "+path)

	switch {
	case path == "/users/@me":
		w.Write([]byte(`{"id":"100","username":"bot"}`))
	case path == "/channels/1":
		w.Write([]byte(`{"id":"1","type":0,"guild_id":"5"}`))
	case path == "/channels/3":
		w.Write([]byte(`{"id":"3","type":11,"guild_id":"5","parent_id":"1"}`))
	case path == "/channels/1/webhooks" && r.Method == http.MethodGet:
		w.Write([]byte(`[{"id":"7","type":1,"token":"other","user":{"id":"200"}}]`))
	case path == "/channels/1/webhooks" && r.Method == http.MethodPost:
		f.created++
		w.Write([]byte(`{"id":"8","type":1,"token":"secret","user":{"id":"100"}}`))
	case strings.HasPrefix(path, "/webhooks/8/secret"):
		if f.deleted {
			f.deleted = false
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code":10015,"message":"Unknown Webhook"}`))
			return
		}
		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &body)
		f.executions = append(f.executions, r)
		f.bodies = append(f.bodies, body)
		w.Write([]byte(`{"id":"9","channel_id":"1","content":"hi"}`))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"code":0,"message":"not found"}`))
	}
}

func newPersonaTestProvider(t *testing.T, features model.Features) (*DiscordProvider, *fakeWebhookDiscord) {
	t.Helper()

	fake := &fakeWebhookDiscord{}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	base := server.URL + api.Path + "/"
	endpoints := []*string{&api.EndpointChannels, &api.EndpointWebhooks, &api.EndpointMe}
	old := []string{api.EndpointChannels, api.EndpointWebhooks, api.EndpointMe}
	api.EndpointChannels = base + "channels/"
	api.EndpointWebhooks = base + "webhooks/"
	api.EndpointMe = base + "users/@me"
	t.Cleanup(func() {
		for i, e := range endpoints {
			*e = old[i]
		}
	})

	p := NewDiscordProvider("app", nil, staticFeatureProvider{features}, nil, NewWebhookCache(), state.New("Bot token"))
	return p, fake
}

func TestCreatePersonaMessage(t *testing.T) {
	p, fake := newPersonaTestProvider(t, model.Features{Personas: true})
	persona := provider.Persona{Name: "Narrator", AvatarURL: "https://example.com/a.png"}

	msg, err := p.CreatePersonaMessage(context.Background(), 1, persona, api.SendMessageData{Content: "hi"})
	require.NoError(t, err)
	assert.Equal(t, discord.MessageID(9), msg.ID)

	// The webhook of the other bot isn't used, a new one is made.
	assert.Equal(t, 1, fake.created)
	require.Len(t, fake.executions, 1)
	assert.Equal(t, "true", fake.executions[0].URL.Query().Get("wait"))
	assert.Equal(t, "Narrator", fake.bodies[0]["username"])
	assert.Equal(t, "https://example.com/a.png", fake.bodies[0]["avatar_url"])
	assert.Equal(t, "hi", fake.bodies[0]["content"])

	// The second message reuses the cached webhook.
	fake.requests = nil
	_, err = p.CreatePersonaMessage(context.Background(), 1, persona, api.SendMessageData{Content: "hi"})
	require.NoError(t, err)
	assert.NotContains(t, fake.requests, "GET /channels/1/webhooks")
	assert.Equal(t, 1, fake.created)
}

func TestCreatePersonaMessageInThread(t *testing.T) {
	p, fake := newPersonaTestProvider(t, model.Features{Personas: true})

	_, err := p.CreatePersonaMessage(context.Background(), 3, provider.Persona{Name: "Narrator"}, api.SendMessageData{Content: "hi"})
	require.NoError(t, err)
	assert.Contains(t, fake.requests, "GET /channels/1/webhooks")
	require.Len(t, fake.executions, 1)
	assert.Equal(t, "3", fake.executions[0].URL.Query().Get("thread_id"))
}

func TestCreatePersonaMessageRecreatesDeletedWebhook(t *testing.T) {
	p, fake := newPersonaTestProvider(t, model.Features{Personas: true})
	persona := provider.Persona{Name: "Narrator"}

	_, err := p.CreatePersonaMessage(context.Background(), 1, persona, api.SendMessageData{Content: "hi"})
	require.NoError(t, err)

	fake.deleted = true
	_, err = p.CreatePersonaMessage(context.Background(), 1, persona, api.SendMessageData{Content: "hi"})
	require.NoError(t, err)
	assert.Equal(t, 2, fake.created)
	assert.Len(t, fake.executions, 2)
}

func TestCreatePersonaMessageRequiresFeature(t *testing.T) {
	p, fake := newPersonaTestProvider(t, model.Features{})

	_, err := p.CreatePersonaMessage(context.Background(), 1, provider.Persona{Name: "Narrator"}, api.SendMessageData{Content: "hi"})
	require.ErrorContains(t, err, "requires premium")
	assert.Empty(t, fake.requests)
}
