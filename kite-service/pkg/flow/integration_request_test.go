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

func executeIntegrationBlockWith(t *testing.T, nodeType FlowNodeType, data FlowNodeData, integrationProvider provider.IntegrationProvider, httpProvider provider.HTTPProvider) (*FlowContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{
			Discord:     &provider.MockDiscordProvider{},
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
	assert.Empty(t, integrations["roblox"].AuthorizeURLFor("Secret-PublicPart"))
}

// executePopcatBlock runs a Popcat block with Popcat enabled.
func executePopcatBlock(t *testing.T, nodeType FlowNodeType, settings map[string]any, httpProvider *redirectCheckingHTTPProvider) (*FlowContext, error) {
	return executeIntegrationBlockWith(t, nodeType, FlowNodeData{Fields: settings},
		&integrationTestProvider{choices: map[string]bool{"popcat": true}}, httpProvider)
}

func TestPopcatSearch(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{"error":false,"message":{"name":"The Octocat","followers":"20000"}}`}
	c, err := executePopcatBlock(t, "action_popcat_search",
		map[string]any{"popcat_endpoint": "github", "popcat_text": "octo cat"}, httpProvider)
	require.NoError(t, err)

	assert.Equal(t, "GET", httpProvider.req.Method)
	assert.Equal(t, "https://api.popcat.xyz/v2/github/octo%20cat", httpProvider.req.URL.String())
	// The result is the message, without the wrapper.
	assert.Equal(t, "The Octocat", c.GetNodeResult("1").Object()["name"].String())
}

func TestPopcatFunGet(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{"error":false,"message":{"answer":"Yes."}}`}
	c, err := executePopcatBlock(t, "action_popcat_fun_get",
		map[string]any{"popcat_endpoint": "8ball"}, httpProvider)
	require.NoError(t, err)

	assert.Equal(t, "https://api.popcat.xyz/v2/8ball", httpProvider.req.URL.String())
	assert.Equal(t, "Yes.", c.GetNodeResult("1").Object()["answer"].String())
}

func TestPopcatTextTransformQuery(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{"error":false,"message":{"text":"hi"}}`}
	_, err := executePopcatBlock(t, "action_popcat_text_transform",
		map[string]any{"popcat_endpoint": "decode", "popcat_text": "01101000 01101001"}, httpProvider)
	require.NoError(t, err)
	assert.Equal(t, "01101000 01101001", httpProvider.req.URL.Query().Get("binary"))
}

// Popcat reports some errors with a 200 and "error": true.
func TestPopcatErrorFlag(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{"error":true,"message":{"error":"User not found"}}`}
	_, err := executePopcatBlock(t, "action_popcat_search",
		map[string]any{"popcat_endpoint": "github", "popcat_text": "nobody"}, httpProvider)
	assert.ErrorContains(t, err, "Popcat returned an error: User not found")
}

func TestPopcatErrorFlagWithMessage(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{"error":true,"message":"Location not found"}`}
	_, err := executePopcatBlock(t, "action_popcat_search",
		map[string]any{"popcat_endpoint": "weather", "popcat_text": "nowhere"}, httpProvider)
	assert.ErrorContains(t, err, "Popcat returned an error: Location not found")
}

// Text to Morse code sets the error flag when it works.
func TestPopcatErrorFlagWithResult(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{"error":true,"message":{"morse":"... --- ..."}}`}
	c, err := executePopcatBlock(t, "action_popcat_text_transform",
		map[string]any{"popcat_endpoint": "texttomorse", "popcat_text": "SOS"}, httpProvider)
	require.NoError(t, err)
	assert.Equal(t, "... --- ...", c.GetNodeResult("1").Object()["morse"].String())
}

// Image endpoints aren't requested, Discord loads the URL.
func TestPopcatImageCreate(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{}
	c, err := executePopcatBlock(t, "action_popcat_image_create", map[string]any{
		"popcat_endpoint": "drake",
		"popcat_text":     "Kite & HTTP blocks",
		"popcat_text_2":   "Popcat blocks",
		// Not sent by the drake meme.
		"popcat_image": "https://example.com/a.png",
	}, httpProvider)
	require.NoError(t, err)

	assert.Nil(t, httpProvider.req)
	assert.Equal(t, "https://api.popcat.xyz/v2/drake?text1=Kite+%26+HTTP+blocks&text2=Popcat+blocks", c.GetNodeResult("1").String())
}

func TestPopcatImageURLTooLong(t *testing.T) {
	_, err := executePopcatBlock(t, "action_popcat_image_create", map[string]any{
		"popcat_endpoint": "sadcat",
		"popcat_text":     strings.Repeat("a", 2100),
	}, &redirectCheckingHTTPProvider{})
	assert.ErrorContains(t, err, "at most 2048 work")
}

func TestPopcatOptionalField(t *testing.T) {
	c, err := executePopcatBlock(t, "action_popcat_image_create", map[string]any{
		"popcat_endpoint": "discord_message",
		"popcat_text":     "hi",
		"popcat_text_2":   "kite",
		"popcat_image":    "https://example.com/a.png",
	}, &redirectCheckingHTTPProvider{})
	require.NoError(t, err)
	assert.NotContains(t, c.GetNodeResult("1").String(), "color=")
}

func TestPopcatInvalidSettings(t *testing.T) {
	for _, tt := range []struct {
		settings map[string]any
		err      string
	}{
		{map[string]any{}, "popcat_endpoint is required"},
		{map[string]any{"popcat_endpoint": "biden"}, "unknown value for popcat_endpoint: biden"},
		// It isn't a template.
		{map[string]any{"popcat_endpoint": "{{'8ball'}}"}, "unknown value for popcat_endpoint"},
		{map[string]any{"popcat_endpoint": "github"}, "popcat_text is required"},
		{map[string]any{"popcat_endpoint": "github", "popcat_text": "../8ball"}, "invalid value for path parameter"},
	} {
		httpProvider := &redirectCheckingHTTPProvider{body: `{"error":false,"message":{}}`}
		_, err := executePopcatBlock(t, "action_popcat_search", tt.settings, httpProvider)
		assert.ErrorContains(t, err, tt.err, tt.settings)
		assert.Nil(t, httpProvider.req, tt.settings)
	}
}

// Popcat is opt-in, as its blocks send it what users enter.
func TestPopcatNotEnabled(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{}
	_, err := executeIntegrationBlockWith(t, "action_popcat_fun_get",
		FlowNodeData{Fields: map[string]any{"popcat_endpoint": "joke"}},
		&integrationTestProvider{}, httpProvider)
	assert.ErrorContains(t, err, "Popcat isn't enabled")
	assert.Nil(t, httpProvider.req)
}
