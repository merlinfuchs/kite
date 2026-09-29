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
)

type integrationTestProvider struct {
	credentials map[string]string
}

func (p *integrationTestProvider) Credential(ctx context.Context, integrationID string) (string, error) {
	if c, ok := p.credentials[integrationID]; ok {
		return c, nil
	}
	return "", provider.ErrNotFound
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
	integrations["test_api"] = Integration{ID: "test_api", Name: "Test API", BaseURL: "https://api.example.com/", Auth: auth}
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
			Integration: &integrationTestProvider{credentials: credentials},
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
	assert.ErrorContains(t, err, "Test API isn't connected")
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
	assert.ErrorContains(t, err, "Test API isn't connected")
	assert.False(t, errors.Is(err, provider.ErrNotFound))
}

func TestCookieAPIQRCode(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{body: `{"success":true,"url":"https://images.cookie-api.com/qr-codes/1.png"}`}

	c, err := executeIntegrationBlock(t, "action_cookie_api_qr_code_create",
		FlowNodeData{Fields: map[string]any{"qr_code_data": "https://kite.onl", "qr_code_border": "2"}},
		map[string]string{"cookie_api": "k3y"}, httpProvider)
	require.NoError(t, err)

	assert.Equal(t, "POST", httpProvider.req.Method)
	assert.Equal(t, "https://api.cookie-api.com/api/images/qr-code", httpProvider.req.URL.String())
	assert.Equal(t, "k3y", httpProvider.req.Header.Get("Authorization"))
	body, _ := io.ReadAll(httpProvider.req.Body)
	assert.JSONEq(t, `{"data":"https://kite.onl","border":2}`, string(body))
	assert.Equal(t, "https://images.cookie-api.com/qr-codes/1.png", c.GetNodeResult("1").Object()["url"].String())
}
