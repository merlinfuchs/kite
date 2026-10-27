package discordapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeState answers every request with result and remembers the requests.
type fakeState struct {
	// Methods the tests don't use panic.
	store.AppStateStore
	store.AppStateManager

	result []byte
	reqs   []provider.DiscordAPIRequest
}

func (f *fakeState) AppState(ctx context.Context, appID string) (store.AppStateStore, error) {
	return f, nil
}

func (f *fakeState) AppDiscordAPIRequest(ctx context.Context, req provider.DiscordAPIRequest) ([]byte, error) {
	f.reqs = append(f.reqs, req)
	return f.result, nil
}

type fakeLogStore struct {
	store.LogStore

	entries []model.LogEntry
}

func (f *fakeLogStore) CreateLogEntry(ctx context.Context, entry model.LogEntry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func serve(t *testing.T, h *DiscordAPIHandler, token *model.APIToken, body string) (int, string, json.RawMessage) {
	t.Helper()

	f := handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		c.APIToken = token
		return handler.TypedWithBody(h.HandleDiscordAPIRequest)(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/discord-api", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	f.ServeHTTP(rec, req)

	var res struct {
		Data struct {
			Result json.RawMessage `json:"result"`
		} `json:"data"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res.Error.Code, res.Data.Result
}

const (
	getChannel  = `{"operation":"get_channel","path_params":{"channel_id":"1"}}`
	sendMessage = `{"operation":"create_message","path_params":{"channel_id":"1"},"body":{"content":"Hi"}}`
)

func TestDiscordAPIRequestScopes(t *testing.T) {
	read := &model.APIToken{Name: "Claude", Scopes: []string{model.APITokenScopeDiscordRead}}
	write := &model.APIToken{Name: "Claude", Scopes: []string{model.APITokenScopeDiscordWrite}}
	kite := &model.APIToken{Name: "Claude", Scopes: []string{model.APITokenScopeKiteWrite}}

	tests := []struct {
		name       string
		token      *model.APIToken
		body       string
		wantStatus int
		wantCode   string
	}{
		{"read gets", read, getChannel, http.StatusOK, ""},
		{"read sends", read, sendMessage, http.StatusForbidden, "missing_api_token_scope"},
		{"write sends", write, sendMessage, http.StatusOK, ""},
		{"kite only", kite, getChannel, http.StatusForbidden, "missing_api_token_scope"},
		{"login", nil, getChannel, http.StatusForbidden, "api_token_required"},
		{"blocked", write, `{"operation":"prune_guild","path_params":{"guild_id":"1"}}`, http.StatusBadRequest, "invalid_discord_api_request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &fakeState{}
			h := NewDiscordAPIHandler(state, &fakeLogStore{})
			status, code, _ := serve(t, h, tt.token, tt.body)
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantCode, code)
			assert.Equal(t, status == http.StatusOK, len(state.reqs) == 1, "request sent")
		})
	}
}

func TestDiscordAPIRequestLogsWrites(t *testing.T) {
	token := &model.APIToken{Name: "Claude", Scopes: []string{model.APITokenScopeDiscordWrite}}
	state := &fakeState{result: []byte(`{"id":"2"}`)}
	logs := &fakeLogStore{}
	h := NewDiscordAPIHandler(state, logs)

	_, _, result := serve(t, h, token, getChannel)
	assert.JSONEq(t, `{"id":"2"}`, string(result))
	assert.Empty(t, logs.entries, "reads aren't logged")

	serve(t, h, token, sendMessage)
	require.Len(t, logs.entries, 1)
	assert.Equal(t, "app", logs.entries[0].AppID)
	assert.Contains(t, logs.entries[0].Message, "create_message (POST /channels/1/messages)")
	assert.Contains(t, logs.entries[0].Message, `"Claude"`)

	req := state.reqs[1]
	assert.JSONEq(t, `{"content":"Hi"}`, string(req.Body))
	assert.Equal(t, `Kite API token "Claude"`, string(req.Reason))
}

func TestDiscordAPIRequestEmptyResult(t *testing.T) {
	token := &model.APIToken{Scopes: []string{model.APITokenScopeDiscordWrite}}
	h := NewDiscordAPIHandler(&fakeState{}, &fakeLogStore{})
	_, _, result := serve(t, h, token, `{"operation":"delete_message","path_params":{"channel_id":"1","message_id":"2"}}`)
	assert.Equal(t, "null", string(result))
}
