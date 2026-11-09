package flow

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// operationsTestProvider answers Discord API requests by path, for calls that
// run at the same time.
type operationsTestProvider struct {
	provider.MockDiscordProvider

	mu        sync.Mutex
	responses map[string]string
	paths     []string
}

func (p *operationsTestProvider) APIRequest(ctx context.Context, req provider.DiscordAPIRequest) ([]byte, error) {
	time.Sleep(10 * time.Millisecond)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paths = append(p.paths, req.Method+" "+req.Path)
	return []byte(p.responses[req.Path]), nil
}

func runJavaScriptOperations(t *testing.T, code string, p *operationsTestProvider, maxCredits int) (*FlowContext, error) {
	c := javaScriptTestContext(t, FlowProviders{
		Discord:     p,
		Log:         &provider.MockLogProvider{},
		Integration: &integrationTestProvider{},
	}, maxCredits, eval.Env{})
	node := &CompiledFlowNode{ID: "js_1", Type: FlowNodeTypeActionJavaScriptRun, Data: FlowNodeData{JavaScriptData: &JavaScriptData{Code: code}}}
	return c, node.Execute(c)
}

func TestJavaScriptRunReadsDiscord(t *testing.T) {
	p := &operationsTestProvider{responses: map[string]string{
		"/channels/5":                  `{"id":"5","name":"general"}`,
		"/guilds/7/members/8":          `{"nick":"Kite"}`,
		"/channels/5/messages?limit=2": `[{"id":"1"},{"id":"2"}]`,
	}}
	c, err := runJavaScriptOperations(t, `
const [channel, member, messages] = await Promise.all([
  discord.getChannel({channel_id: "5"}),
  discord.getGuildMember({guild_id: "7", user_id: "8"}),
  discord.listMessages({channel_id: "5", limit: 2}),
])
return [channel.name, member.nick, messages.length]`, p, 1000)
	require.NoError(t, err)
	assert.JSONEq(t, `["general","Kite",2]`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))
	assert.ElementsMatch(t, []string{"GET /channels/5", "GET /guilds/7/members/8", "GET /channels/5/messages?limit=2"}, p.paths)
	// The block's credits, Discord reads are free.
	assert.Equal(t, 10, c.CreditsUsed())
}

func TestJavaScriptRunOperationErrors(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"write", `await discord.createMessage({channel_id: "5", content: "hi"})`, "Error: discord.createMessage isn't available"},
		{"unknown", `await discord.doSomething()`, "Error: discord.doSomething isn't available"},
		{"missing parameter", `await discord.getChannel({})`, "Error: discord.getChannel: missing path parameter channel_id"},
		{"unknown parameter", `await discord.getChannel({channel_id: "5", foo: 1})`, "Error: discord.getChannel: it has no parameter foo"},
		{"not an object", `await discord.getChannel("5")`, "Error: discord.getChannel: the argument must be an object of its parameters"},
		{"path traversal", `await discord.getChannel({channel_id: "../guilds/1"})`, "Error: discord.getChannel: invalid value for parameter channel_id: must be an ID"},
		{"not enabled", `await erlc.getServer()`, "Error: erlc.getServer: ER:LC isn't enabled, enable it in the app's integrations"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &operationsTestProvider{}
			_, err := runJavaScriptOperations(t, tt.code, p, 1000)
			assert.ErrorContains(t, err, tt.want)
			assert.Empty(t, p.paths)
		})
	}
}

func TestJavaScriptRunOperationCredits(t *testing.T) {
	operations["discord"]["get_paid_thing"] = operation{Integration: "discord", ID: "get_paid_thing", Method: "GET", Path: "/paid", Credits: 3}
	t.Cleanup(func() { delete(operations["discord"], "get_paid_thing") })

	c, err := runJavaScriptOperations(t, `await discord.getPaidThing(); return 1`, &operationsTestProvider{}, 1000)
	require.NoError(t, err)
	assert.Equal(t, 13, c.CreditsUsed())

	// A call over the credit limit fails, like a block, and the code can
	// catch that like an Error handler block can.
	p := &operationsTestProvider{}
	c, err = runJavaScriptOperations(t, `
const errors = []
for (let i = 0; i < 2; i++) {
  try { await discord.getPaidThing() } catch (e) { errors.push(e.message) }
}
return errors`, p, 15)
	require.NoError(t, err)
	assert.JSONEq(t, `["discord.getPaidThing: credits limit exceeded: 15"]`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))
	assert.Len(t, p.paths, 1)
	// Only the call that was sent is billed.
	assert.Equal(t, 13, c.CreditsUsed())

	// So is a call with an argument it can't send.
	c, err = runJavaScriptOperations(t, `try { await discord.getPaidThing({foo: 1}) } catch {} return 1`, p, 1000)
	require.NoError(t, err)
	assert.Equal(t, 10, c.CreditsUsed())
}

func TestJavaScriptRunOperationsLimit(t *testing.T) {
	c := javaScriptTestContext(t, FlowProviders{Discord: &operationsTestProvider{}, Log: &provider.MockLogProvider{}}, 1000, eval.Env{})
	// The block itself is one operation.
	c.MaxOperations = 2
	node := &CompiledFlowNode{ID: "js_1", Type: FlowNodeTypeActionJavaScriptRun, Data: FlowNodeData{JavaScriptData: &JavaScriptData{Code: `
await discord.getChannel({channel_id: "5"})
try { await discord.getChannel({channel_id: "5"}) } catch (e) { return e.message }`}}}
	require.NoError(t, node.Execute(c))
	assert.Equal(t, "discord.getChannel: operations limit exceeded: 2", c.GetNodeResult("js_1").String())
}

func TestJavaScriptRunOperationsInTestRuns(t *testing.T) {
	c := javaScriptTestContext(t, FlowProviders{
		Discord:     &operationsTestProvider{responses: map[string]string{"/channels/5": `{"name":"general"}`}},
		Log:         &provider.MockLogProvider{},
		Integration: &integrationTestProvider{choices: map[string]bool{"erlc": true}},
	}, 1000, eval.Env{})
	c.Test = NewTestRun(nil)
	node := &CompiledFlowNode{ID: "js_1", Type: FlowNodeTypeActionJavaScriptRun, Data: FlowNodeData{JavaScriptData: &JavaScriptData{Code: `
const channel = await discord.getChannel({channel_id: "5"})
try { await erlc.getServer() } catch (e) { return [channel.name, e.message] }`}}}
	require.NoError(t, node.Execute(c))
	assert.JSONEq(t, `["general","erlc.getServer: test runs don't send requests outside of Discord"]`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))
}

func toJSON(t *testing.T, v any) string {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

// Code calls the operations of an integration through a namespace that
// pkg/js/prelude.js defines, named after it in camel case.
func TestEveryIntegrationHasANamespace(t *testing.T) {
	for integration := range operations {
		namespace := integration
		for strings.Contains(namespace, "_") {
			i := strings.Index(namespace, "_")
			namespace = namespace[:i] + strings.ToUpper(namespace[i+1:i+2]) + namespace[i+2:]
		}
		c, err := runJavaScriptOperations(t, "return typeof "+namespace, &operationsTestProvider{}, 1000)
		require.NoError(t, err)
		assert.Equal(t, "object", c.GetNodeResult("js_1").String(), integration)
	}
}

func TestOperationRequestBody(t *testing.T) {
	path, body, err := operationRequest(operations["discord"]["create_message"], json.RawMessage(`{"channel_id":"5","body":{"content":"Hi"}}`))
	require.NoError(t, err)
	assert.Equal(t, "/channels/5/messages", path)
	assert.JSONEq(t, `{"content":"Hi"}`, string(body))

	// Bodies that are lists work the same.
	_, body, err = operationRequest(operations["discord"]["bulk_update_guild_roles"], json.RawMessage(`{"guild_id":"7","body":[{"id":"8","position":1}]}`))
	require.NoError(t, err)
	assert.JSONEq(t, `[{"id":"8","position":1}]`, string(body))

	_, _, err = operationRequest(operations["discord"]["create_message"], json.RawMessage(`{"channel_id":"5","content":"Hi"}`))
	assert.EqualError(t, err, "it has no parameter content")
	_, _, err = operationRequest(operations["discord"]["get_channel"], json.RawMessage(`{"channel_id":"5","body":{}}`))
	assert.EqualError(t, err, "it has no parameter body")
}

// Code passes the body and the audit log reason of writes next to the path
// and query parameters, so no operation can have parameters of those names.
func TestOperationParamsAreNotReserved(t *testing.T) {
	for _, ops := range operations {
		for _, op := range ops {
			for _, p := range slices.Concat(op.PathParams, op.QueryParams) {
				assert.NotContains(t, []string{"body", "reason"}, p.Name, "%s.%s", op.Integration, op.ID)
			}
		}
	}
}
