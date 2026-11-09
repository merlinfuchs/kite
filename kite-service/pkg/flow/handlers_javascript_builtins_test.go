package flow

import (
	"context"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
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

// codeVariableProvider keeps the values of stored variables in memory.
type codeVariableProvider struct {
	mu     sync.Mutex
	scoped map[string]bool
	values map[string]thing.Thing
}

func (p *codeVariableProvider) key(id string, scope null.String) string {
	return id + "/" + scope.String
}

func (p *codeVariableProvider) VariableScoped(ctx context.Context, id string) (bool, error) {
	return p.scoped[id], nil
}

func (p *codeVariableProvider) UpdateVariable(ctx context.Context, id string, scope null.String, op provider.VariableOperation, value thing.Thing) (thing.Thing, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if op == provider.VariableOperationIncrement {
		current, ok := p.values[p.key(id, scope)]
		if !ok {
			current = thing.NewInt(0)
		}
		value = current.Add(value)
	}
	p.values[p.key(id, scope)] = value
	return value, nil
}

func (p *codeVariableProvider) Variable(ctx context.Context, id string, scope null.String) (thing.Thing, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	v, ok := p.values[p.key(id, scope)]
	if !ok {
		return thing.Null, provider.ErrNotFound
	}
	return v, nil
}

func (p *codeVariableProvider) DeleteVariable(ctx context.Context, id string, scope null.String) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.values, p.key(id, scope))
	return nil
}

func (p *codeVariableProvider) VariableValues(ctx context.Context, id string, opts provider.VariableListOptions) ([]provider.VariableValue, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var res []provider.VariableValue
	for key, v := range p.values {
		varID, scope, _ := strings.Cut(key, "/")
		if varID == id && strings.HasPrefix(scope, opts.ScopePrefix) {
			res = append(res, provider.VariableValue{Scope: null.NewString(scope, scope != ""), Value: v, UpdatedAt: time.Unix(100, 0)})
		}
	}
	slices.SortFunc(res, func(a, b provider.VariableValue) int { return strings.Compare(a.Scope.String, b.Scope.String) })
	return res, nil
}

func runJavaScriptVariables(t *testing.T, code string, p *codeVariableProvider, maxCredits int) (*FlowContext, error) {
	c := javaScriptTestContext(t, FlowProviders{Variable: p, Log: &provider.MockLogProvider{}}, maxCredits, eval.Env{})
	node := &CompiledFlowNode{ID: "js_1", Type: FlowNodeTypeActionJavaScriptRun, Data: FlowNodeData{JavaScriptData: &JavaScriptData{
		Code: code,
		Variables: []JavaScriptVariable{
			{Name: "xp", VariableID: "var_xp"},
			{Name: "motd", VariableID: "var_motd"},
			{Name: "unpicked"},
		},
	}}}
	return c, node.Execute(c)
}

func TestJavaScriptRunVariables(t *testing.T) {
	p := &codeVariableProvider{scoped: map[string]bool{"var_xp": true}, values: map[string]thing.Thing{}}
	c, err := runJavaScriptVariables(t, `
const before = await variables.get({name: "motd"})
await variables.set({name: "motd", value: {text: "Hi", n: [1, 2]}})
await Promise.all([1, 2, 3].map(() => variables.increment({name: "xp", scope: "1"})))
await variables.increment({name: "xp", scope: "2", by: 10})
const xp = await variables.get({name: "xp", scope: "1"})
const top = await variables.list({name: "xp"})
await variables.delete({name: "xp", scope: "2"})
return {before, motd: await variables.get({name: "motd", scope: "ignored"}), xp, top, after: await variables.list({name: "xp"})}`, p, 1000)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"before": null,
		"motd": {"text": "Hi", "n": [1, 2]},
		"xp": 3,
		"top": [{"scope": "1", "value": 3, "updated_at": 100}, {"scope": "2", "value": 10, "updated_at": 100}],
		"after": [{"scope": "1", "value": 3, "updated_at": 100}]
	}`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))
	// The block, and set, 4 increments, 2 lists and a delete. Gets are free.
	assert.Equal(t, 10+8, c.CreditsUsed())
}

func TestJavaScriptRunVariableErrors(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"not declared", `await variables.get({name: "coins"})`, "variables.get: the block has no stored variable named coins, add it under Stored variables"},
		{"not picked", `await variables.get({name: "unpicked"})`, "variables.get: no variable is picked for unpicked under Stored variables of the block"},
		{"no scope", `await variables.increment({name: "xp"})`, "variables.increment: xp has a value per scope, pass the scope, like a user ID"},
		{"no value", `await variables.set({name: "motd"})`, "variables.set: missing parameter value"},
		{"no name", `await variables.get({})`, "variables.get: missing parameter name"},
		{"sort", `await variables.list({name: "xp", sort: "random"})`, "variables.list: invalid value for parameter sort: must be value_desc, value_asc, updated_desc or updated_asc"},
		{"limit", `await variables.list({name: "xp", limit: 500})`, "variables.list: invalid value for parameter limit: must be from 1 to 100"},
		{"scope type", `await variables.get({name: "xp", scope: 5})`, "variables.get: invalid value for parameter scope"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &codeVariableProvider{scoped: map[string]bool{"var_xp": true}, values: map[string]thing.Thing{}}
			c, err := runJavaScriptVariables(t, tt.code, p, 1000)
			assert.ErrorContains(t, err, tt.want)
			assert.Empty(t, p.values)
			assert.Equal(t, 10, c.CreditsUsed())
		})
	}
}
