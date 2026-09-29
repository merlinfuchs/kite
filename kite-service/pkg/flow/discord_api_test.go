package flow

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscordAPIOperations(t *testing.T) {
	for _, id := range []string{
		"create_message",
		"list_messages",
		"bulk_delete_messages",
		"prune_guild",
		"leave_guild",
		"update_my_guild_member",
		"get_my_application",
	} {
		assert.Contains(t, discordAPIOperations, id)
	}

	// Managed by Kite or not authenticated with the bot token.
	for _, id := range []string{
		"update_my_application",
		"bulk_set_application_commands",
		"create_application_emoji",
		"update_my_user",
		"create_interaction_response",
		"execute_webhook",
		"create_guild_sticker",
	} {
		assert.NotContains(t, discordAPIOperations, id)
	}
}

func TestDiscordAPIPath(t *testing.T) {
	str := func(s string) thing.Thing { return thing.NewString(s) }

	tests := map[string]struct {
		op    string
		path  map[string]thing.Thing
		query map[string]thing.Thing
		want  string
		err   string
	}{
		"path params": {
			op:   "get_guild_member",
			path: map[string]thing.Thing{"guild_id": str("1"), "user_id": str(" 2 ")},
			want: "/guilds/1/members/2",
		},
		"discord object as ID": {
			op:   "get_channel",
			path: map[string]thing.Thing{"channel_id": thing.NewDiscordChannel(discord.Channel{ID: 3})},
			want: "/channels/3",
		},
		"escaped string param": {
			op: "add_my_message_reaction",
			path: map[string]thing.Thing{
				"channel_id": str("1"), "message_id": str("2"), "emoji_name": str("👍"),
			},
			want: "/channels/1/messages/2/reactions/%F0%9F%91%8D/@me",
		},
		"query": {
			op:    "list_messages",
			path:  map[string]thing.Thing{"channel_id": str("1")},
			query: map[string]thing.Thing{"limit": str("5"), "before": str("9")},
			want:  "/channels/1/messages?before=9&limit=5",
		},
		"traversal in ID": {
			op:   "get_channel",
			path: map[string]thing.Thing{"channel_id": str("1/../../guilds/2")},
			err:  "invalid value for parameter channel_id: must be an ID",
		},
		"traversal in string param": {
			op:   "invite_resolve",
			path: map[string]thing.Thing{"code": str("..")},
			err:  "invalid value for path parameter code",
		},
		"slash in string param": {
			op:   "invite_resolve",
			path: map[string]thing.Thing{"code": str("abc/../x")},
			err:  "invalid value for path parameter code",
		},
		"mention as ID": {
			op:   "get_channel",
			path: map[string]thing.Thing{"channel_id": str("<#1>")},
			err:  "invalid value for parameter channel_id: must be an ID",
		},
		"empty ID": {
			op:   "get_channel",
			path: map[string]thing.Thing{"channel_id": thing.Null},
			err:  "invalid value for parameter channel_id: must be an ID",
		},
		"missing path param": {
			op:  "get_channel",
			err: "missing path parameter channel_id",
		},
		"unknown path param": {
			op:   "get_channel",
			path: map[string]thing.Thing{"channel_id": str("1"), "guild_id": str("2")},
			err:  "unknown path parameter guild_id",
		},
		"unknown query param": {
			op:    "get_channel",
			path:  map[string]thing.Thing{"channel_id": str("1")},
			query: map[string]thing.Thing{"limit": str("5")},
			err:   "unknown query parameter limit",
		},
		"invalid integer": {
			op:    "list_messages",
			path:  map[string]thing.Thing{"channel_id": str("1")},
			query: map[string]thing.Thing{"limit": str("")},
			err:   `invalid value for parameter limit: strconv.ParseInt: parsing "": invalid syntax`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			op, ok := discordAPIOperations[tt.op]
			require.True(t, ok)

			path, err := discordAPIPath(op, tt.path, tt.query)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, path)
		})
	}
}

type discordAPITestProvider struct {
	provider.MockDiscordProvider

	req provider.DiscordAPIRequest
	err error
}

func (p *discordAPITestProvider) APIRequest(ctx context.Context, req provider.DiscordAPIRequest) ([]byte, error) {
	p.req = req
	if p.err != nil {
		return nil, p.err
	}
	return []byte(`{"id":"5","count":2,"items":[{"name":"a"}]}`), nil
}

func executeDiscordAPIRequest(t *testing.T, p provider.DiscordProvider, data FlowNodeData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{
			Discord: p,
			Log:     &provider.MockLogProvider{},
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	defer c.Cancel()

	node := &CompiledFlowNode{ID: "1", Type: FlowNodeTypeActionDiscordAPIRequest, Data: data}
	return node.Execute(c)
}

func TestExecuteDiscordAPIRequest(t *testing.T) {
	p := &discordAPITestProvider{}
	err := executeDiscordAPIRequest(t, p, FlowNodeData{
		AuditLogReason: "because {{1 + 1}}",
		DiscordAPIRequestData: &DiscordAPIRequestData{
			Operation:  "create_message",
			PathParams: []HTTPRequestDataKeyValue{{Key: "channel_id", Value: "{{1 + 2}}"}},
			BodyJSON: json.RawMessage(`{
				"content": "Hi {{1 + 1}} ",
				"empty": "",
				"id": 123456789012345678901,
				"nonce": "{{1 + 1}}",
				"list": ["{{true}}", "{{[1, 2]}}"]
			}`),
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "POST", p.req.Method)
	assert.Equal(t, "/channels/3/messages", p.req.Path)
	assert.Equal(t, api.AuditLogReason("because 2"), p.req.Reason)
	assert.JSONEq(t, `{
		"content": "Hi 2 ",
		"empty": "",
		"id": 123456789012345678901,
		"nonce": 2,
		"list": [true, [1, 2]]
	}`, string(p.req.Body))
}

func TestExecuteDiscordAPIRequestErrors(t *testing.T) {
	tests := map[string]struct {
		data     DiscordAPIRequestData
		provider error
		err      string
	}{
		"unknown endpoint": {
			data: DiscordAPIRequestData{Operation: "update_my_application"},
			err:  "unknown Discord API endpoint: update_my_application",
		},
		"body on endpoint without one": {
			data: DiscordAPIRequestData{
				Operation:  "get_channel",
				PathParams: []HTTPRequestDataKeyValue{{Key: "channel_id", Value: "1"}},
				BodyJSON:   json.RawMessage(`{}`),
			},
			err: "the get_channel endpoint doesn't take a body",
		},
		"failed request": {
			data: DiscordAPIRequestData{
				Operation:  "get_channel",
				PathParams: []HTTPRequestDataKeyValue{{Key: "channel_id", Value: "1"}},
			},
			provider: errors.New("Discord 404 error: Unknown Channel"),
			err:      "Discord 404 error: Unknown Channel",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := &discordAPITestProvider{err: tt.provider}
			err := executeDiscordAPIRequest(t, p, FlowNodeData{DiscordAPIRequestData: &tt.data})
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestDiscordAPIResult(t *testing.T) {
	res, err := discordAPIResult([]byte(`{"code":"abc","max_age":3600,"channel":{"id":"1"},"items":[{"name":"a"}]}`))
	require.NoError(t, err)

	obj := res.Object()
	assert.Equal(t, "abc", obj["code"].String())
	assert.Equal(t, int64(3600), obj["max_age"].Int())
	assert.Equal(t, "1", obj["channel"].Object()["id"].String())
	assert.Equal(t, "a", obj["items"].Array()[0].Object()["name"].String())

	// Results are stored with the flow's state, e.g. across a sleep.
	raw, err := json.Marshal(res)
	require.NoError(t, err)
	var restored thing.Thing
	require.NoError(t, json.Unmarshal(raw, &restored))
	assert.Equal(t, res, restored)

	res, err = discordAPIResult(nil)
	require.NoError(t, err)
	assert.True(t, res.IsNil())
}
