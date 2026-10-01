package flow

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type integrationTestProvider struct {
	credentials map[string]string
	choices     map[string]bool
}

func (p *integrationTestProvider) Credential(ctx context.Context, integrationID string) (string, error) {
	if c, ok := p.credentials[integrationID]; ok {
		return c, nil
	}
	return "", provider.ErrNotFound
}

func (p *integrationTestProvider) Choice(ctx context.Context, integrationID string) (null.Bool, error) {
	enabled, ok := p.choices[integrationID]
	return null.NewBool(enabled, ok), nil
}

type redirectCheckingHTTPProvider struct {
	req       *http.Request
	redirects bool
	status    int
	body      string
}

func (p *redirectCheckingHTTPProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	p.redirects = true
	return p.HTTPRequestWithoutRedirects(ctx, req)
}

func (p *redirectCheckingHTTPProvider) HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error) {
	p.req = req
	status := p.status
	if status == 0 {
		status = 200
	}
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Body: io.NopCloser(strings.NewReader(p.body))}, nil
}

// withTestIntegration adds an integration with an API key and a block that
// sends a request to it, for the duration of a test.
func withTestIntegration(t *testing.T, auth IntegrationAuth) {
	availability := "opt_in"
	if auth.Type == "none" {
		availability = "default"
	}
	integrations["test_api"] = Integration{ID: "test_api", Name: "Test API", BaseURL: "https://api.example.com/", Auth: auth, Availability: availability}
	blockDefinitions["action_test_api_thing_get"] = blockDefinition{
		Type:     "action_test_api_thing_get",
		Credits:  intPtr(2),
		Requires: []string{"test_api"},
		Run:      blockRequest{Kind: "request", Integration: "test_api", Method: "GET", Path: "/things/{id}"},
		Fields:   []blockField{{Name: "thing_id", In: "path", Target: "id", Type: "string"}},
		Result: &struct {
			Thing string `json:"thing"`
			List  bool   `json:"list"`
		}{},
	}
	t.Cleanup(func() {
		delete(integrations, "test_api")
		delete(blockDefinitions, "action_test_api_thing_get")
	})
}

func intPtr(v int) *int { return &v }

func executeIntegrationBlock(t *testing.T, nodeType FlowNodeType, data FlowNodeData, credentials map[string]string, httpProvider provider.HTTPProvider) (*FlowContext, error) {
	// Entering a credential enables the integration.
	choices := map[string]bool{}
	for id := range credentials {
		choices[id] = true
	}
	return executeIntegrationBlockWith(t, nodeType, data, &integrationTestProvider{credentials: credentials, choices: choices}, httpProvider)
}

// integrationTestDiscordProvider has a bot token, for blocks that send it.
type integrationTestDiscordProvider struct {
	provider.MockDiscordProvider
}

func (p *integrationTestDiscordProvider) BotToken() string {
	return "b0t-t0ken"
}

func executeIntegrationBlockWith(t *testing.T, nodeType FlowNodeType, data FlowNodeData, integrationProvider provider.IntegrationProvider, httpProvider provider.HTTPProvider) (*FlowContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{
			Discord:     &integrationTestDiscordProvider{},
			HTTP:        httpProvider,
			Log:         &provider.MockLogProvider{},
			Integration: integrationProvider,
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	node := &CompiledFlowNode{ID: "1", Type: nodeType, Data: data}
	return c, node.Execute(c)
}

func TestIntegrationRequestHeaderAuth(t *testing.T) {
	withTestIntegration(t, IntegrationAuth{Type: "header", Name: "Authorization", Prefix: "Bearer "})
	httpProvider := &redirectCheckingHTTPProvider{body: `{"name":"thing"}`}

	c, err := executeIntegrationBlock(t, "action_test_api_thing_get",
		FlowNodeData{Fields: map[string]any{"thing_id": "abc"}},
		map[string]string{"test_api": "k3y"}, httpProvider)
	require.NoError(t, err)

	assert.False(t, httpProvider.redirects, "follows redirects")
	assert.Equal(t, "https://api.example.com/things/abc", httpProvider.req.URL.String())
	assert.Equal(t, "Bearer k3y", httpProvider.req.Header.Get("Authorization"))
	assert.Equal(t, "thing", c.GetNodeResult("1").Object()["name"].String())
}

func TestIntegrationRequestQueryAuth(t *testing.T) {
	withTestIntegration(t, IntegrationAuth{Type: "query", Name: "api_key"})
	httpProvider := &redirectCheckingHTTPProvider{body: `{}`}

	_, err := executeIntegrationBlock(t, "action_test_api_thing_get",
		FlowNodeData{Fields: map[string]any{"thing_id": "abc"}},
		map[string]string{"test_api": "k3y"}, httpProvider)
	require.NoError(t, err)
	assert.Equal(t, "k3y", httpProvider.req.URL.Query().Get("api_key"))
}

func TestIntegrationRequestNotConnected(t *testing.T) {
	withTestIntegration(t, IntegrationAuth{Type: "header", Name: "Authorization"})
	httpProvider := &redirectCheckingHTTPProvider{}

	_, err := executeIntegrationBlock(t, "action_test_api_thing_get",
		FlowNodeData{Fields: map[string]any{"thing_id": "abc"}}, nil, httpProvider)
	assert.ErrorContains(t, err, "Test API isn't enabled")
	assert.Nil(t, httpProvider.req)
}

func TestIntegrationRequestErrorRedactsCredential(t *testing.T) {
	withTestIntegration(t, IntegrationAuth{Type: "query", Name: "api_key"})
	// The key ends where the error is cut, so it's only redacted if that
	// happens first.
	body := strings.Repeat("ü", 296) + "long-k3y"
	httpProvider := &redirectCheckingHTTPProvider{status: 401, body: body}

	_, err := executeIntegrationBlock(t, "action_test_api_thing_get",
		FlowNodeData{Fields: map[string]any{"thing_id": "abc"}},
		map[string]string{"test_api": "long-k3y"}, httpProvider)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Test API returned")
	assert.NotContains(t, err.Error(), "long")
	assert.True(t, utf8.ValidString(err.Error()))
}

// Blocks written in Go that need an integration fail before they run when
// the app didn't connect it.
func TestCustomBlockNeedsIntegration(t *testing.T) {
	withTestIntegration(t, IntegrationAuth{Type: "header", Name: "Authorization"})
	block := blockDefinitions[FlowNodeTypeActionLog]
	blockDefinitions[FlowNodeTypeActionLog] = blockDefinition{Type: block.Type, Requires: []string{"test_api"}, Run: block.Run}
	t.Cleanup(func() { blockDefinitions[FlowNodeTypeActionLog] = block })

	_, err := executeIntegrationBlock(t, FlowNodeTypeActionLog, FlowNodeData{LogMessage: "hi"}, nil, &redirectCheckingHTTPProvider{})
	assert.ErrorContains(t, err, "Test API isn't enabled")
	assert.False(t, errors.Is(err, provider.ErrNotFound))
}

func TestCookieAPIBlocks(t *testing.T) {
	tests := []struct {
		nodeType  FlowNodeType
		data      FlowNodeData
		response  string
		method    string
		url       string
		body      string
		resultKey string
		want      string
	}{
		{
			nodeType:  "action_cookie_api_transcript_create",
			data:      FlowNodeData{ChannelTarget: "123", Fields: map[string]any{"transcript_title": "Ticket"}},
			response:  `{"success":true,"url":"https://www.cookie-api.com/transcripts/1/2-ticket.html"}`,
			method:    "POST",
			url:       "https://api.cookie-api.com/api/transcript?channel_id=123",
			body:      `{"bot_token":"b0t-t0ken","title":"Ticket"}`,
			resultKey: "url",
			want:      "https://www.cookie-api.com/transcripts/1/2-ticket.html",
		},
		{
			nodeType:  "action_cookie_api_card_create",
			data:      FlowNodeData{Fields: map[string]any{"card_data": `{"card":{"width":"145","height":"40","bg":"#FF0000","bg_type":"color"},"elements":[{"type":"text","text":"Hi {{ 'there' }}"}]}`}},
			response:  `{"success":true,"url":"https://cards.cookie-api.com/card-builder/1/2.png"}`,
			method:    "POST",
			url:       "https://api.cookie-api.com/api/cards/card-builder/build",
			body:      `{"card":{"width":"145","height":"40","bg":"#FF0000","bg_type":"color"},"elements":[{"type":"text","text":"Hi there"}]}`,
			resultKey: "url",
			want:      "https://cards.cookie-api.com/card-builder/1/2.png",
		},
		{
			nodeType:  "action_cookie_api_qr_code_create",
			data:      FlowNodeData{Fields: map[string]any{"qr_code_data": "https://kite.onl", "qr_code_border": "2"}},
			response:  `{"success":true,"url":"https://images.cookie-api.com/qr-codes/1.png"}`,
			method:    "POST",
			url:       "https://api.cookie-api.com/api/images/qr-code",
			body:      `{"data":"https://kite.onl","border":2}`,
			resultKey: "url",
			want:      "https://images.cookie-api.com/qr-codes/1.png",
		},
		{
			nodeType:  "action_cookie_api_captcha_create",
			data:      FlowNodeData{Fields: map[string]any{"captcha_provider": "Cloudflare", "captcha_color": "#FFFFFF"}},
			response:  `{"success":true,"captcha_id":"2725738690","url":"https://api.cookie-api.com/public/captcha?code=abc"}`,
			method:    "POST",
			url:       "https://api.cookie-api.com/api/security/captcha/create?captcha_provider=Cloudflare&color=%23FFFFFF",
			resultKey: "captcha_id",
			want:      "2725738690",
		},
		{
			nodeType:  "action_cookie_api_captcha_get",
			data:      FlowNodeData{Fields: map[string]any{"captcha_id": "2725738690"}},
			response:  `{"success":true,"captcha_id":"2725738690","solved":"YES","solved_at":"1739110529"}`,
			method:    "GET",
			url:       "https://api.cookie-api.com/api/security/captcha/get-captcha?captcha_id=2725738690",
			resultKey: "solved",
			want:      "YES",
		},
		{
			nodeType:  "action_cookie_api_minecraft_user_get",
			data:      FlowNodeData{Fields: map[string]any{"minecraft_code": "12345"}},
			response:  `{"success":true,"edition":"JAVA","player_id":"1d25b1dc57e14bb9bd93f574f75cb4d0","player_name":"The_Tea_Cookie"}`,
			method:    "GET",
			url:       "https://api.cookie-api.com/api/minecraft/get-user?code=12345",
			resultKey: "player_name",
			want:      "The_Tea_Cookie",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.nodeType), func(t *testing.T) {
			httpProvider := &redirectCheckingHTTPProvider{body: tt.response}

			c, err := executeIntegrationBlock(t, tt.nodeType, tt.data,
				map[string]string{"cookie_api": "k3y"}, httpProvider)
			require.NoError(t, err)

			assert.Equal(t, tt.method, httpProvider.req.Method)
			assert.Equal(t, tt.url, httpProvider.req.URL.String())
			assert.Equal(t, "k3y", httpProvider.req.Header.Get("Authorization"))
			if tt.body == "" {
				assert.Nil(t, httpProvider.req.Body)
			} else {
				body, _ := io.ReadAll(httpProvider.req.Body)
				assert.JSONEq(t, tt.body, string(body))
			}
			assert.Equal(t, tt.want, c.GetNodeResult("1").Object()[tt.resultKey].String())
		})
	}
}

func TestIntegrationRequestErrorRedactsBotToken(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{status: 400, body: `{"success":false,"message":"invalid token b0t-t0ken"}`}

	_, err := executeIntegrationBlock(t, "action_cookie_api_transcript_create",
		FlowNodeData{ChannelTarget: "123"},
		map[string]string{"cookie_api": "k3y"}, httpProvider)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "b0t-t0ken")
}

func TestCookieAPIRejectsInvalidCard(t *testing.T) {
	_, err := executeIntegrationBlock(t, "action_cookie_api_card_create",
		FlowNodeData{Fields: map[string]any{"card_data": `["not", "an", "object"]`}},
		map[string]string{"cookie_api": "k3y"}, &redirectCheckingHTTPProvider{})
	assert.ErrorContains(t, err, "must be a JSON object")
}

func TestCookieAPIRejectsUnknownOption(t *testing.T) {
	_, err := executeIntegrationBlock(t, "action_cookie_api_captcha_create",
		FlowNodeData{Fields: map[string]any{"captcha_provider": "{{ 'hcaptcha' }}"}},
		map[string]string{"cookie_api": "k3y"}, &redirectCheckingHTTPProvider{})
	assert.ErrorContains(t, err, "must be one of Cloudflare, Google")
}

// Integrations without a credential are on by default, until the app turns
// them off.
func TestCustomBlockNeedsEnabledIntegration(t *testing.T) {
	withTestIntegration(t, IntegrationAuth{Type: "none"})
	block := blockDefinitions[FlowNodeTypeActionLog]
	blockDefinitions[FlowNodeTypeActionLog] = blockDefinition{Type: block.Type, Requires: []string{"test_api"}, Run: block.Run}
	t.Cleanup(func() { blockDefinitions[FlowNodeTypeActionLog] = block })

	data := FlowNodeData{LogMessage: "hi"}
	_, err := executeIntegrationBlockWith(t, FlowNodeTypeActionLog, data, &integrationTestProvider{}, &redirectCheckingHTTPProvider{})
	require.NoError(t, err)

	_, err = executeIntegrationBlockWith(t, FlowNodeTypeActionLog, data,
		&integrationTestProvider{choices: map[string]bool{"test_api": false}}, &redirectCheckingHTTPProvider{})
	assert.ErrorContains(t, err, "Test API isn't enabled")
}

func TestIntegrationEnabled(t *testing.T) {
	always := Integration{Availability: "always", Auth: IntegrationAuth{Type: "discord_bot"}}
	byDefault := Integration{Availability: "default", Auth: IntegrationAuth{Type: "none"}}
	optIn := Integration{Availability: "opt_in", Auth: IntegrationAuth{Type: "header"}}

	assert.True(t, always.Enabled(null.BoolFrom(false)))
	assert.True(t, byDefault.Enabled(null.Bool{}))
	assert.False(t, byDefault.Enabled(null.BoolFrom(false)))
	assert.False(t, optIn.Enabled(null.Bool{}))
	assert.True(t, optIn.Enabled(null.BoolFrom(true)))
	assert.False(t, optIn.Enabled(null.BoolFrom(false)))
}

// Integrations that need a credential can't be on before the app connected
// them.
func TestIntegrationAvailability(t *testing.T) {
	for _, integration := range Integrations() {
		assert.Contains(t, []string{"always", "default", "opt_in"}, integration.Availability, integration.ID)
		if integration.NeedsCredential() {
			assert.Equal(t, "opt_in", integration.Availability, integration.ID)
		}
	}
}

// A disabled integration keeps its credential, but its blocks don't run.
func TestIntegrationRequestDisabled(t *testing.T) {
	withTestIntegration(t, IntegrationAuth{Type: "header", Name: "Authorization"})
	httpProvider := &redirectCheckingHTTPProvider{}

	_, err := executeIntegrationBlockWith(t, "action_test_api_thing_get",
		FlowNodeData{Fields: map[string]any{"thing_id": "abc"}},
		&integrationTestProvider{
			credentials: map[string]string{"test_api": "k3y"},
			choices:     map[string]bool{"test_api": false},
		}, httpProvider)
	assert.ErrorContains(t, err, "Test API isn't enabled")
	assert.Nil(t, httpProvider.req)
}

// configureTestERLC sets Kite's key and ID at ER:LC for the duration of a
// test.
func configureTestERLC(t *testing.T) {
	require.NoError(t, ConfigureIntegration("erlc", "k1te-k3y", "1234"))
	t.Cleanup(func() { _ = ConfigureIntegration("erlc", "", "") })
}

// executeERLCBlock runs an ER:LC block, and checks it sent the server key
// and Kite's key.
func executeERLCBlock(t *testing.T, nodeType FlowNodeType, data FlowNodeData, httpProvider *redirectCheckingHTTPProvider) *FlowContext {
	configureTestERLC(t)
	c, err := executeIntegrationBlock(t, nodeType, data, map[string]string{"erlc": "s3cret-server"}, httpProvider)
	require.NoError(t, err)
	assert.Equal(t, "s3cret-server", httpProvider.req.Header.Get("server-key"))
	assert.Equal(t, "k1te-k3y", httpProvider.req.Header.Get("Authorization"))
	return c
}

func TestERLCServerGet(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{"Name":"API Test","CurrentPlayers":3,"Players":[],"Queue":[]}`}
	c := executeERLCBlock(t, "action_erlc_server_get",
		FlowNodeData{Fields: map[string]any{"include_players": "true", "include_queue": "{{ 'true' }}"}}, httpProvider)

	assert.Equal(t, "GET", httpProvider.req.Method)
	assert.Equal(t, "https://api.erlc.gg/v2/server?Players=true&Queue=true", httpProvider.req.URL.String())
	assert.Equal(t, "3", c.GetNodeResult("1").Object()["CurrentPlayers"].String())
}

func TestERLCCommandRun(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{"message":"Success"}`}
	executeERLCBlock(t, "action_erlc_command_run",
		FlowNodeData{Fields: map[string]any{"erlc_command": ":h {{ 'Hello' }}"}}, httpProvider)

	assert.Equal(t, "POST", httpProvider.req.Method)
	assert.Equal(t, "https://api.erlc.gg/v2/server/command", httpProvider.req.URL.String())
	body, _ := io.ReadAll(httpProvider.req.Body)
	assert.JSONEq(t, `{"command":":h Hello"}`, string(body))
}

// Self-hosted instances have no key at ER:LC, and send none.
func TestERLCWithoutKiteKey(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{}`}

	_, err := executeIntegrationBlock(t, "action_erlc_server_get", FlowNodeData{},
		map[string]string{"erlc": "s3cret-server"}, httpProvider)
	require.NoError(t, err)
	assert.Empty(t, httpProvider.req.Header.Values("Authorization"))
}

func TestERLCErrorRedactsKiteKey(t *testing.T) {
	configureTestERLC(t)
	httpProvider := &redirectCheckingHTTPProvider{status: 403, body: `{"code":2003,"message":"invalid global API key k1te-k3y"}`}

	_, err := executeIntegrationBlock(t, "action_erlc_server_get", FlowNodeData{},
		map[string]string{"erlc": "s3cret-server"}, httpProvider)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "k1te-k3y")
}

func TestIntegrationAuthorizeURL(t *testing.T) {
	configureTestERLC(t)
	erlc, _ := GetIntegration("erlc")

	assert.Equal(t, "https://api.erlc.gg/server-owners/server/PublicPart/authorize/1234", erlc.AuthorizeURLFor("Secret-PublicPart"))
	assert.Empty(t, erlc.AuthorizeURLFor("NoDash"))
	assert.Empty(t, erlc.AuthorizeURLFor("Secret-"))
	assert.Empty(t, integrations["cookie_api"].AuthorizeURLFor("Secret-PublicPart"))
}
