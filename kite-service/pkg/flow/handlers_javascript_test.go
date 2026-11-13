package flow

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// javaScriptTestContext is the context of a Run JavaScript block in tests.
func javaScriptTestContext(t *testing.T, providers FlowProviders, maxCredits int, env eval.Env) *FlowContext {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		providers,
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: maxCredits},
		eval.NewContext(env),
		nil,
	)
	t.Cleanup(c.Cancel)
	return c
}

func executeJavaScript(t *testing.T, data JavaScriptData, env eval.Env, results ...map[string]thing.Thing) (thing.Thing, []string, error) {
	logs := &recordingLogProvider{}
	c := javaScriptTestContext(t, FlowProviders{Log: logs}, 1000, env)

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
		Code: `console.log("types", typeof input.n, typeof input.text, Array.isArray(input.list)); console.log("second"); console.warn("warn")
return {sum: input.list.reduce((a, b) => a + b, input.n), text: input.text + "!", empty: input.empty}`,
		Inputs: []JavaScriptInput{
			{Name: "n", Value: "{{1 + 2}}"},
			{Name: "text", Value: "hi {{1 + 2}}"},
			{Name: "list", Value: "{{[1, 2, 3]}}"},
			{Name: "empty"},
		},
	}, eval.Env{"user": eval.NewUserEnv(user)})
	require.NoError(t, err)
	assert.Equal(t, []string{"types number string true\nsecond", "warn"}, logs)
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
		"channel_1": thing.NewDiscordChannel(discord.Channel{ID: 11, Name: "general"}),
		"role_1":    thing.NewDiscordRole(discord.Role{ID: 12, Name: "Mod"}),
		"guild_1":   thing.NewDiscordGuild(discord.Guild{ID: 13, Name: "Kite"}),
		"message_1": thing.NewDiscordMessage(discord.Message{ID: 14, Content: "hi", Author: user}),
		"http_1":    thing.NewHTTPResponse(thing.HTTPResponseValue{Status: "200 OK", StatusCode: 200, Body: []byte(`{"a": [1]}`)}),
	}
	res, _, err := executeJavaScript(t, JavaScriptData{
		Code: `return [input.user.display_name, input.user.mention, input.user.id, typeof input.user.role_names,
input.member.nick, input.member.role_ids, input.members[0].username, input.name, input.http.status_code, input.http.data.a, input.http.body,
input.channel.name, input.role.name, input.guild.name, input.message.content, input.message.author.username]`,
		Inputs: []JavaScriptInput{
			{Name: "user", Value: "{{user}}"},
			{Name: "member", Value: "{{result('member_1')}}"},
			{Name: "members", Value: "{{result('members_1')}}"},
			{Name: "name", Value: "{{user.username}}"},
			{Name: "http", Value: "{{result('http_1')}}"},
			{Name: "channel", Value: "{{result('channel_1')}}"},
			{Name: "role", Value: "{{result('role_1')}}"},
			{Name: "guild", Value: "{{result('guild_1')}}"},
			{Name: "message", Value: "{{result('message_1')}}"},
		},
	}, eval.Env{"user": eval.NewUserEnv(user)}, results)
	require.NoError(t, err)
	got := toJSON(t, res.JSONValue())
	assert.JSONEq(t, `["Kite", "<@123456789012345678>", "123456789012345678", "undefined",
"Nick", ["42"], "member", "kite", 200, [1], "{\"a\": [1]}",
"general", "Mod", "Kite", "hi", "kite"]`, got)
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
