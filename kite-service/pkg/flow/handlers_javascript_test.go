package flow

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeJavaScript(t *testing.T, data JavaScriptData, env eval.Env, results ...map[string]thing.Thing) (thing.Thing, []string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	logs := &recordingLogProvider{}
	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{Log: logs},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(env),
		nil,
	)
	t.Cleanup(c.Cancel)

	for _, r := range results {
		for id, result := range r {
			c.StoreNodeResult(&CompiledFlowNode{ID: id}, result)
		}
	}

	node := &CompiledFlowNode{ID: "js_1", Type: FlowNodeTypeActionJavaScriptRun, Data: FlowNodeData{JavaScriptData: &data}}
	err := node.Execute(c)
	return c.GetNodeResult("js_1"), logs.messages, err
}

func TestJavaScriptRunInputs(t *testing.T) {
	user := discord.User{ID: 123456789012345678, Username: "kite", DisplayName: "Kite"}
	res, logs, err := executeJavaScript(t, JavaScriptData{
		Code: `console.log("types", typeof input.n, typeof input.text, Array.isArray(input.list))
return {sum: input.list.reduce((a, b) => a + b, input.n), text: input.text + "!", empty: input.empty}`,
		Inputs: []JavaScriptInput{
			{Name: "n", Value: "{{1 + 2}}"},
			{Name: "text", Value: "hi {{1 + 2}}"},
			{Name: "list", Value: "{{[1, 2, 3]}}"},
			{Name: "empty"},
		},
	}, eval.Env{"user": eval.NewUserEnv(user)})
	require.NoError(t, err)
	assert.Equal(t, []string{"types number string true"}, logs)
	obj := res.Object()
	assert.Equal(t, int64(9), obj["sum"].Int())
	assert.Equal(t, "hi 3!", obj["text"].String())
	assert.Equal(t, "", obj["empty"].String())
}

func TestJavaScriptRunDiscordInputs(t *testing.T) {
	user := discord.User{ID: 123456789012345678, Username: "kite", DisplayName: "Kite"}
	member := discord.Member{User: discord.User{ID: 223456789012345678, Username: "member"}, Nick: "Nick", RoleIDs: []discord.RoleID{42}}
	results := map[string]thing.Thing{
		"member_1":  thing.NewDiscordMember(member),
		"members_1": thing.NewArray([]thing.Thing{thing.NewDiscordMember(member)}),
		"http_1":    thing.NewHTTPResponse(thing.HTTPResponseValue{Status: "200 OK", StatusCode: 200, Body: []byte(`{"a": [1]}`)}),
	}
	res, _, err := executeJavaScript(t, JavaScriptData{
		Code: `return [input.user.display_name, input.user.mention, input.user.id, typeof input.user.role_names,
input.member.nick, input.member.role_ids, input.members[0].username, input.name, input.http.status_code, input.http.data.a, input.http.body]`,
		Inputs: []JavaScriptInput{
			{Name: "user", Value: "{{user}}"},
			{Name: "member", Value: "{{result('member_1')}}"},
			{Name: "members", Value: "{{result('members_1')}}"},
			{Name: "name", Value: "{{user.username}}"},
			{Name: "http", Value: "{{result('http_1')}}"},
		},
	}, eval.Env{"user": eval.NewUserEnv(user)}, results)
	require.NoError(t, err)
	got, err := json.Marshal(res.JSONValue())
	require.NoError(t, err)
	assert.JSONEq(t, `["Kite", "<@123456789012345678>", "123456789012345678", "undefined",
"Nick", ["42"], "member", "kite", 200, [1], "{\"a\": [1]}"]`, string(got))
}

func TestJavaScriptRunPlaceholdersInCodeAreText(t *testing.T) {
	res, _, err := executeJavaScript(t, JavaScriptData{Code: `return "{{user.username}}"`}, eval.Env{"user": eval.NewUserEnv(discord.User{Username: "kite"})})
	require.NoError(t, err)
	assert.Equal(t, "{{user.username}}", res.String())
}

func TestJavaScriptRunError(t *testing.T) {
	_, _, err := executeJavaScript(t, JavaScriptData{Code: "const a = 1\nthrow new Error('no ' + a)"}, eval.Env{})
	assert.ErrorContains(t, err, "Error: no 1 (line 2)")
}

func TestJavaScriptDataValidate(t *testing.T) {
	ok := JavaScriptData{Code: "return 1", Inputs: []JavaScriptInput{{Name: "a"}, {Name: "_b2"}}}
	assert.NoError(t, ok.Validate())
	for _, data := range []JavaScriptData{
		{},
		{Code: "return 1", Inputs: []JavaScriptInput{{Name: ""}}},
		{Code: "return 1", Inputs: []JavaScriptInput{{Name: "2a"}}},
		{Code: "return 1", Inputs: []JavaScriptInput{{Name: "a-b"}}},
		{Code: "return 1", Inputs: []JavaScriptInput{{Name: "a"}, {Name: "a"}}},
	} {
		assert.Error(t, data.Validate(), "%+v", data)
	}
}
