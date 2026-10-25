package engine

import (
	"context"
	"encoding/json"
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

type fakeVariableStore struct {
	store.VariableStore
	variables []*model.Variable
	loads     int
}

func (f *fakeVariableStore) VariablesByAppWithoutTotals(ctx context.Context, appID string) ([]*model.Variable, error) {
	f.loads++
	return f.variables, nil
}

func TestVariableProviderVariableScoped(t *testing.T) {
	variables := &fakeVariableStore{variables: []*model.Variable{
		{ID: "scoped", Scoped: true},
		{ID: "unscoped", Scoped: false},
	}}
	p := NewVariableProvider("app", variables, nil)

	for id, want := range map[string]bool{"scoped": true, "unscoped": false, "unknown": false} {
		scoped, err := p.VariableScoped(context.Background(), id)
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if scoped != want {
			t.Errorf("%s: got scoped %v, want %v", id, scoped, want)
		}
	}
	if variables.loads != 1 {
		t.Errorf("loaded variables %d times, want 1", variables.loads)
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

	p := NewDiscordProvider("app", nil, nil, nil, state.New("Bot token"))

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

	// The path of a request that never reaches Discord can have a token.
	server.Close()
	_, err = p.APIRequest(context.Background(), provider.DiscordAPIRequest{
		Method: http.MethodPost,
		Path:   "/webhooks/1/secret-token",
	})
	if err == nil || strings.Contains(err.Error(), "secret-token") {
		t.Errorf("unreachable: got %v", err)
	}
}

func TestDecisionProviderSendsQuestion(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer key" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		w.Write([]byte(`{"model":"typesafe/jev-1.14","answers":{"answer":{"type":"choice","choice":"spam","probabilities":{"spam":0.8,"other":0.2}}},"usage":{"input_tokens":10,"output_tokens":1}}`))
	}))
	defer server.Close()

	p := NewDecisionProvider(server.Client(), "key")
	p.url = server.URL

	decision, err := p.Decide(context.Background(), provider.DecideOpts{
		Model:    "typesafe/jev-1.14",
		Type:     provider.DecisionTypeChoice,
		Question: "What is it?",
		Input:    "Free nitro",
		Options:  map[string]string{"spam": "Advertising", "other": ""},
	})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Choice != "spam" || decision.Confidence != 0.8 {
		t.Errorf("decision = %+v", decision)
	}

	question := body["questions"].(map[string]any)["answer"].(map[string]any)
	criteria := question["criteria"].(map[string]any)
	if body["state"] != "Free nitro" || question["type"] != "choice" || question["instructions"] != "What is it?" {
		t.Errorf("body = %v", body)
	}
	if criteria["spam"] != "Advertising" || criteria["other"] != nil {
		t.Errorf("criteria = %v", criteria)
	}
}

func TestDecisionProviderReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":400,"message":"Too many options"}}`))
	}))
	defer server.Close()

	p := NewDecisionProvider(server.Client(), "key")
	p.url = server.URL

	_, err := p.Decide(context.Background(), provider.DecideOpts{Type: provider.DecisionTypeNoul})
	if err == nil || !strings.Contains(err.Error(), "status 400: Too many options") {
		t.Errorf("err = %v", err)
	}
}

func TestDecisionProviderHidesKeyErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		w.Write([]byte(`{"error":{"code":402,"message":"Insufficient credits"}}`))
	}))
	defer server.Close()

	p := NewDecisionProvider(server.Client(), "key")
	p.url = server.URL

	_, err := p.Decide(context.Background(), provider.DecideOpts{Type: provider.DecisionTypeNoul})
	if err == nil || strings.Contains(err.Error(), "credits") {
		t.Errorf("err = %v", err)
	}
}
