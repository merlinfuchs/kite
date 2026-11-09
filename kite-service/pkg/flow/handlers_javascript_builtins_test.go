package flow

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// codeHTTPProvider answers every request with status and body, and keeps
// what it was sent.
type codeHTTPProvider struct {
	status int
	body   string

	reqs    []*http.Request
	reqBody string
}

func (p *codeHTTPProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	p.reqs = append(p.reqs, req)
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		p.reqBody = string(b)
	}
	return &http.Response{
		StatusCode: p.status,
		Status:     http.StatusText(p.status),
		Header:     http.Header{"X-Thing": {"a", "b"}},
		Body:       io.NopCloser(strings.NewReader(p.body)),
	}, nil
}

func (p *codeHTTPProvider) HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error) {
	return p.HTTPRequest(ctx, req)
}

func runJavaScriptBuiltins(t *testing.T, code string, httpProvider *codeHTTPProvider, maxCredits int) (*FlowContext, error) {
	c := javaScriptTestContext(t, FlowProviders{
		HTTP:   httpProvider,
		Secret: &secretTestProvider{values: map[string]string{"TOKEN": "s3cret-t0ken"}},
		Log:    &provider.MockLogProvider{},
	}, maxCredits, eval.Env{})
	node := &CompiledFlowNode{ID: "js_1", Type: FlowNodeTypeActionJavaScriptRun, Data: FlowNodeData{JavaScriptData: &JavaScriptData{Code: code}}}
	return c, node.Execute(c)
}

func TestJavaScriptRunFetch(t *testing.T) {
	p := &codeHTTPProvider{status: 201, body: `{"id":5}`}
	c, err := runJavaScriptBuiltins(t, `
const res = await fetch("https://api.example.com/things?token={{secrets.TOKEN}}", {
  method: "POST",
  headers: {Authorization: "Bearer {{secrets.TOKEN}}", "Content-Type": "application/json"},
  body: JSON.stringify({name: "{{secrets.TOKEN}}", other: "{{user.id}}"}),
})
return [res.status, res.statusText, res.headers.get("x-thing"), await res.json()]`, p, 1000)
	require.NoError(t, err)
	assert.JSONEq(t, `[201,"Created","a,b",{"id":5}]`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))

	require.Len(t, p.reqs, 1)
	req := p.reqs[0]
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "https://api.example.com/things?token=s3cret-t0ken", req.URL.String())
	assert.Equal(t, "Bearer s3cret-t0ken", req.Header.Get("Authorization"))
	// Only secrets are filled in, code has no placeholders.
	assert.JSONEq(t, `{"name":"s3cret-t0ken","other":"{{user.id}}"}`, p.reqBody)
	assert.Equal(t, 13, c.CreditsUsed())
}

func TestJavaScriptRunHTTPRequest(t *testing.T) {
	// Like fetch, a response that isn't 2xx resolves.
	p := &codeHTTPProvider{status: 404, body: "not here"}
	c, err := runJavaScriptBuiltins(t, `
const res = await http.request({url: "https://example.com"})
return [res.status, res.body, (await fetch("https://example.com")).ok]`, p, 1000)
	require.NoError(t, err)
	assert.JSONEq(t, `[404,"not here",false]`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))

	// Text unless the code says otherwise, like fetch.
	p = &codeHTTPProvider{status: 200}
	_, err = runJavaScriptBuiltins(t, `await fetch("https://example.com", {method: "POST", body: "hi"})`, p, 1000)
	require.NoError(t, err)
	assert.Equal(t, "text/plain;charset=UTF-8", p.reqs[0].Header.Get("Content-Type"))
}

func TestJavaScriptRunHTTPRequestErrors(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"no url", `await http.request({})`, "http.request: missing parameter url"},
		{"not http", `await http.request({url: "file:///etc/passwd"})`, "http.request: invalid value for parameter url: must start with http:// or https://"},
		{"method", `await http.request({url: "https://example.com", method: "TRACE"})`, "http.request: invalid value for parameter method: must be one of GET, POST, PUT, PATCH, DELETE"},
		{"unknown parameter", `await http.request({url: "https://example.com", foo: 1})`, "http.request: it has no parameter foo"},
		{"wrong type", `await http.request({url: "https://example.com", body: {a: 1}})`, "http.request: invalid value for parameter body"},
		{"not an object", `await http.request("https://example.com")`, "http.request: the argument must be an object of its parameters"},
		{"unknown secret", `await fetch("https://example.com/{{secrets.NOPE}}")`, "http.request: the app has no secret named NOPE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &codeHTTPProvider{status: 200}
			c, err := runJavaScriptBuiltins(t, tt.code, p, 1000)
			assert.ErrorContains(t, err, tt.want)
			assert.Empty(t, p.reqs)
			// Only the block is billed.
			assert.Equal(t, 10, c.CreditsUsed())
		})
	}
}

func TestJavaScriptRunHTTPRequestInTestRuns(t *testing.T) {
	c := javaScriptTestContext(t, FlowProviders{HTTP: &codeHTTPProvider{status: 200}, Log: &provider.MockLogProvider{}}, 1000, eval.Env{})
	c.Test = NewTestRun(nil)
	node := &CompiledFlowNode{ID: "js_1", Type: FlowNodeTypeActionJavaScriptRun, Data: FlowNodeData{JavaScriptData: &JavaScriptData{Code: `
try { await fetch("https://example.com") } catch (e) { return e.message }`}}}
	require.NoError(t, node.Execute(c))
	assert.Equal(t, "http.request: test runs don't send requests outside of Discord", c.GetNodeResult("js_1").String())
}

func TestEveryHandlerOperationHasAHandler(t *testing.T) {
	for integration, ops := range operations {
		for id, op := range ops {
			if op.Kind == "handler" {
				assert.Contains(t, operationHandlers, integration+"."+id)
			}
		}
	}
	for name := range operationHandlers {
		integration, id, _ := strings.Cut(name, ".")
		assert.Equal(t, "handler", operations[integration][id].Kind, name)
	}
}
