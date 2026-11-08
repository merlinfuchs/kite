package flow

import (
	"context"
	"encoding/json"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{
			Discord:     p,
			Log:         &provider.MockLogProvider{},
			Integration: &integrationTestProvider{},
		},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: maxCredits},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

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
	operations["discord"]["get_paid_thing"] = operation{Integration: "discord", ID: "get_paid_thing", Method: "GET", Path: "/paid", Policy: operationPolicy{Credits: 3}}
	t.Cleanup(func() { delete(operations["discord"], "get_paid_thing") })

	c, err := runJavaScriptOperations(t, `await discord.getPaidThing(); return 1`, &operationsTestProvider{}, 1000)
	require.NoError(t, err)
	assert.Equal(t, 13, c.CreditsUsed())

	// Running out of credits fails the block even if the code catches it.
	p := &operationsTestProvider{}
	_, err = runJavaScriptOperations(t, `
for (let i = 0; i < 2; i++) {
  try { await discord.getPaidThing() } catch {}
}
return 1`, p, 15)
	assert.ErrorContains(t, err, "credits limit exceeded")
	assert.Len(t, p.paths, 1)
}

func toJSON(t *testing.T, v any) string {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}
