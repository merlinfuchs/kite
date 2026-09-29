package flow

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type secretTestProvider struct {
	values map[string]string
	asked  [][]string
}

func (p *secretTestProvider) Secrets(ctx context.Context, names []string) (map[string]string, error) {
	p.asked = append(p.asked, names)
	res := map[string]string{}
	for _, name := range names {
		if v, ok := p.values[name]; ok {
			res[name] = v
		}
	}
	return res, nil
}

type httpTestProvider struct {
	req *http.Request
	err error
}

func (p *httpTestProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	p.req = req
	if p.err != nil {
		return nil, p.err
	}
	return &http.Response{StatusCode: 200, Status: "200 OK", Body: io.NopCloser(strings.NewReader("{}"))}, nil
}

func executeWithSecrets(t *testing.T, node *CompiledFlowNode, secrets *secretTestProvider, httpProvider *httpTestProvider) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{
			Discord: &provider.MockDiscordProvider{},
			HTTP:    httpProvider,
			Log:     &provider.MockLogProvider{},
			Secret:  secrets,
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)
	return node.Execute(c)
}

func httpNode(data HTTPRequestData) *CompiledFlowNode {
	return &CompiledFlowNode{ID: "1", Type: FlowNodeTypeActionHTTPRequest, Data: FlowNodeData{HTTPRequestData: &data}}
}

func TestHTTPRequestSecrets(t *testing.T) {
	secrets := &secretTestProvider{values: map[string]string{"API_KEY": "s3cr3t", "OTHER": "unused"}}
	httpProvider := &httpTestProvider{}

	err := executeWithSecrets(t, httpNode(HTTPRequestData{
		URL:      "https://example.com/?key={{secrets.API_KEY}}",
		Method:   "POST",
		Headers:  []HTTPRequestDataKeyValue{{Key: "Authorization", Value: "Bearer {{secrets.API_KEY}}"}},
		BodyJSON: []byte(`{"token":"{{secrets.API_KEY}}"}`),
	}), secrets, httpProvider)
	require.NoError(t, err)

	// Only the referenced secret is fetched, once.
	assert.Equal(t, [][]string{{"API_KEY"}}, secrets.asked)
	assert.Equal(t, "s3cr3t", httpProvider.req.URL.Query().Get("key"))
	assert.Equal(t, "Bearer s3cr3t", httpProvider.req.Header.Get("Authorization"))
	body, _ := io.ReadAll(httpProvider.req.Body)
	assert.JSONEq(t, `{"token":"s3cr3t"}`, string(body))
}

func TestHTTPRequestWithoutSecrets(t *testing.T) {
	secrets := &secretTestProvider{}
	err := executeWithSecrets(t, httpNode(HTTPRequestData{URL: "https://example.com"}), secrets, &httpTestProvider{})
	require.NoError(t, err)
	assert.Empty(t, secrets.asked)
}

func TestHTTPRequestUnknownSecret(t *testing.T) {
	err := executeWithSecrets(t, httpNode(HTTPRequestData{URL: "https://example.com/?key={{secrets.MISSING}}"}), &secretTestProvider{}, &httpTestProvider{})
	assert.ErrorContains(t, err, "the app has no secret named MISSING")
}

func TestHTTPRequestRedactsSecrets(t *testing.T) {
	secrets := &secretTestProvider{values: map[string]string{"API_KEY": "s3cr3t value"}}
	httpProvider := &httpTestProvider{err: errors.New(`Get "https://example.com/?key=s3cr3t+value": dial tcp: lookup failed`)}

	err := executeWithSecrets(t, httpNode(HTTPRequestData{URL: "https://example.com/?key={{secrets.API_KEY}}"}), secrets, httpProvider)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "s3cr3t")
	assert.Contains(t, err.Error(), "key=[secret]")
}

// Secrets only work in requests, so they can't be sent in a message or
// written to the logs by accident.
func TestSecretsOnlyInRequests(t *testing.T) {
	secrets := &secretTestProvider{values: map[string]string{"API_KEY": "s3cr3t"}}
	node := &CompiledFlowNode{ID: "1", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "{{secrets.API_KEY}}"}}

	err := executeWithSecrets(t, node, secrets, &httpTestProvider{})
	require.Error(t, err)
	assert.Empty(t, secrets.asked)
}
