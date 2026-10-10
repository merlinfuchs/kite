package flow

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// staticHTTPProvider answers every request with the same response.
type staticHTTPProvider struct {
	req     *http.Request
	status  int
	headers http.Header
	body    string
}

func (p *staticHTTPProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	p.req = req
	return &http.Response{
		StatusCode: p.status,
		Status:     http.StatusText(p.status),
		Header:     p.headers,
		Body:       io.NopCloser(strings.NewReader(p.body)),
	}, nil
}

func (p *staticHTTPProvider) HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error) {
	return p.HTTPRequest(ctx, req)
}

func newTestEvalContext(values map[string]any) eval.Context {
	env := eval.Env{}
	for k, v := range values {
		env[k] = v
	}
	return eval.NewContext(env)
}

func runHTTPNode(t *testing.T, data HTTPRequestData, httpProvider *httpTestProvider) (*FlowContext, error) {
	t.Helper()

	node := httpNode(data)
	ctx := NewContext(
		context.Background(),
		5*time.Second,
		&TestContextData{},
		FlowProviders{HTTP: httpProvider, Secret: &secretTestProvider{}},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		newTestEvalContext(map[string]any{"name": `Jack "the dev"`, "count": 3}),
		nil,
	)
	t.Cleanup(ctx.Cancel)
	return ctx, node.Execute(ctx)
}

func TestHTTPRequestJSONBody(t *testing.T) {
	httpProvider := &httpTestProvider{}
	_, err := runHTTPNode(t, HTTPRequestData{
		URL:      "https://example.com",
		Method:   "POST",
		BodyType: HTTPRequestBodyTypeJSON,
		Body:     `{"content": "Hi {{name}}", "count": {{count}}}`,
	}, httpProvider)
	require.NoError(t, err)

	assert.Equal(t, "application/json", httpProvider.req.Header.Get("Content-Type"))
	body, _ := io.ReadAll(httpProvider.req.Body)
	assert.JSONEq(t, `{"content": "Hi Jack \"the dev\"", "count": 3}`, string(body))
	assert.Equal(t, int64(len(body)), httpProvider.req.ContentLength)
}

// Blocks saved before body types only have body_json, which is still sent.
func TestHTTPRequestLegacyBody(t *testing.T) {
	httpProvider := &httpTestProvider{}
	_, err := runHTTPNode(t, HTTPRequestData{
		URL:      "https://example.com",
		Method:   "POST",
		BodyJSON: []byte(`{"content": "Hi {{name}}"}`),
	}, httpProvider)
	require.NoError(t, err)

	body, _ := io.ReadAll(httpProvider.req.Body)
	assert.JSONEq(t, `{"content": "Hi Jack \"the dev\""}`, string(body))
}

func TestHTTPRequestNoBody(t *testing.T) {
	for _, data := range []HTTPRequestData{
		{URL: "https://example.com"},
		{URL: "https://example.com", BodyJSON: []byte("null")},
		// Turning the JSON body off keeps the old body around, unsent.
		{URL: "https://example.com", BodyType: HTTPRequestBodyTypeNone, Body: `{"a": 1}`},
	} {
		httpProvider := &httpTestProvider{}
		_, err := runHTTPNode(t, data, httpProvider)
		require.NoError(t, err)
		assert.Nil(t, httpProvider.req.Body)
		assert.Empty(t, httpProvider.req.Header.Get("Content-Type"))
	}
}

func TestHTTPRequestContentTypeHeaderWins(t *testing.T) {
	httpProvider := &httpTestProvider{}
	_, err := runHTTPNode(t, HTTPRequestData{
		URL:      "https://example.com",
		Headers:  []HTTPRequestDataKeyValue{{Key: "Content-Type", Value: "application/vnd.api+json"}},
		BodyType: HTTPRequestBodyTypeJSON,
		Body:     `{}`,
	}, httpProvider)
	require.NoError(t, err)
	assert.Equal(t, "application/vnd.api+json", httpProvider.req.Header.Get("Content-Type"))
}

func TestHTTPRequestInvalidBody(t *testing.T) {
	for _, data := range []HTTPRequestData{
		{URL: "https://example.com", BodyType: HTTPRequestBodyTypeJSON, Body: `{"a":}`},
		// Text, form and multipart bodies were never offered.
		{URL: "https://example.com", BodyType: "text", Body: "hi"},
	} {
		httpProvider := &httpTestProvider{}
		_, err := runHTTPNode(t, data, httpProvider)
		require.Error(t, err)
		assert.Nil(t, httpProvider.req, "nothing is sent when the body is invalid")
	}
}

func TestHTTPRequestResponseOptions(t *testing.T) {
	httpProvider := &staticHTTPProvider{
		status:  http.StatusOK,
		headers: http.Header{"X-Total": []string{"7"}},
		body:    `{"items": [{"name": "first"}]}`,
	}

	ctx := NewContext(
		context.Background(),
		5*time.Second,
		&TestContextData{},
		FlowProviders{HTTP: httpProvider, Secret: &secretTestProvider{}},
		FlowContextLimits{},
		newTestEvalContext(nil),
		nil,
	)
	t.Cleanup(ctx.Cancel)

	res, err := ctx.executeHTTPRequest(&HTTPRequestData{
		URL:               "https://example.com",
		ResponseTransform: `{{ response.data().items[0].name + "/" + response.headers["X-Total"] }}`,
	})
	require.NoError(t, err)
	assert.Equal(t, "first/7", res.String())

	// Without a transform the response is the result.
	res, err = ctx.executeHTTPRequest(&HTTPRequestData{URL: "https://example.com"})
	require.NoError(t, err)
	assert.Equal(t, thing.TypeHTTPResponse, res.Type)

	httpProvider.status = http.StatusNotFound
	httpProvider.body = `{"error": "missing"}`

	// Error statuses are passed on unless the block fails on them.
	_, err = ctx.executeHTTPRequest(&HTTPRequestData{URL: "https://example.com"})
	require.NoError(t, err)

	_, err = ctx.executeHTTPRequest(&HTTPRequestData{URL: "https://example.com", FailOnErrorStatus: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Not Found")
	assert.Contains(t, err.Error(), "missing")
}

func TestCheckHTTPResponseStatus(t *testing.T) {
	assert.NoError(t, checkHTTPResponseStatus(thing.HTTPResponseValue{StatusCode: 204}))
	assert.NoError(t, checkHTTPResponseStatus(thing.HTTPResponseValue{StatusCode: 302}))

	err := checkHTTPResponseStatus(thing.HTTPResponseValue{Status: "500 Internal Server Error", StatusCode: 500})
	assert.EqualError(t, err, "request failed with status 500 Internal Server Error")

	// Long bodies are cut without splitting a character.
	err = checkHTTPResponseStatus(thing.HTTPResponseValue{
		Status:     "404 Not Found",
		StatusCode: 404,
		Body:       []byte(strings.Repeat("é", 400)),
	})
	require.Error(t, err)
	assert.True(t, strings.HasSuffix(err.Error(), "…"))
	assert.Less(t, len(err.Error()), 400)
	assert.NotContains(t, err.Error(), "�")
}

func TestTransformHTTPResponseKeepsEnv(t *testing.T) {
	ctx := NewContext(
		context.Background(),
		5*time.Second,
		&TestContextData{},
		FlowProviders{},
		FlowContextLimits{},
		newTestEvalContext(nil),
		nil,
	)
	t.Cleanup(ctx.Cancel)

	resp := thing.HTTPResponseValue{Status: "200 OK", StatusCode: 200, Body: []byte(`{"a": 1}`)}

	res, err := ctx.transformHTTPResponse("  ", resp)
	require.NoError(t, err)
	assert.Equal(t, thing.TypeHTTPResponse, res.Type)

	res, err = ctx.transformHTTPResponse("response.status_code", resp)
	require.NoError(t, err)
	assert.Equal(t, int64(200), res.Int())

	// The response only exists inside the transform.
	_, hasResponse := ctx.EvalCtx.Env["response"]
	assert.False(t, hasResponse)
}
