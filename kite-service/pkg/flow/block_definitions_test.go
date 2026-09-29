package flow

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
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

type blockTestProvider struct {
	provider.MockDiscordProvider

	req      provider.DiscordAPIRequest
	response string
}

func (p *blockTestProvider) APIRequest(ctx context.Context, req provider.DiscordAPIRequest) ([]byte, error) {
	p.req = req
	return []byte(p.response), nil
}

type blockTestContextData struct {
	TestContextData
}

func (d *blockTestContextData) GuildID() discord.GuildID {
	return 5
}

func (d *blockTestContextData) ChannelID() discord.ChannelID {
	return 6
}

func executeBlock(t *testing.T, p provider.DiscordProvider, nodeType FlowNodeType, dataJSON string) (*FlowContext, error) {
	var data FlowNodeData
	require.NoError(t, json.Unmarshal([]byte(dataJSON), &data))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&blockTestContextData{},
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
	t.Cleanup(c.Cancel)

	node := &CompiledFlowNode{ID: "1", Type: nodeType, Data: data}
	return c, node.Execute(c)
}

func TestBlockDefinitionRoleCreate(t *testing.T) {
	p := &blockTestProvider{response: `{"id":"7","name":"Mods","color":16711680}`}
	c, err := executeBlock(t, p, "action_role_create", `{
		"audit_log_reason": "because",
		"name": "Mods",
		"permissions": "8",
		"color": "{{16711680}}",
		"hoist": true
	}`)
	require.NoError(t, err)

	assert.Equal(t, "POST", p.req.Method)
	// The guild the flow runs in.
	assert.Equal(t, "/guilds/5/roles", p.req.Path)
	assert.Equal(t, api.AuditLogReason("because"), p.req.Reason)
	assert.JSONEq(t, `{"name":"Mods","permissions":"8","color":16711680,"hoist":true}`, string(p.req.Body))

	result := c.GetNodeResult("1")
	assert.Equal(t, thing.TypeDiscordRole, result.Type)
	assert.Equal(t, discord.RoleID(7), result.DiscordRole().ID)
}

func TestBlockDefinitionMessageList(t *testing.T) {
	p := &blockTestProvider{response: `[{"id":"8","content":"hi"},{"id":"9","content":"hello"}]`}
	c, err := executeBlock(t, p, "action_message_list", `{
		"channel_target": "3",
		"limit": 5,
		"before": "10"
	}`)
	require.NoError(t, err)

	assert.Equal(t, "GET", p.req.Method)
	assert.Equal(t, "/channels/3/messages?before=10&limit=5", p.req.Path)
	assert.Nil(t, p.req.Body)

	result := c.GetNodeResult("1").Array()
	require.Len(t, result, 2)
	assert.Equal(t, "hi", result[0].DiscordMessage().Content)
}

func TestBlockDefinitionBulkDelete(t *testing.T) {
	for name, ids := range map[string]string{
		"separated":              `"11, 12 13"`,
		"placeholder":            `"{{['11', '12', '13']}}"`,
		"list":                   `["11", "12", "13"]`,
		"list with placeholders": `["{{'11'}}", "12", "13"]`,
	} {
		t.Run(name, func(t *testing.T) {
			p := &blockTestProvider{}
			_, err := executeBlock(t, p, "action_message_bulk_delete", `{"message_ids": `+ids+`}`)
			require.NoError(t, err)

			// The channel the flow runs in.
			assert.Equal(t, "/channels/6/messages/bulk-delete", p.req.Path)
			assert.JSONEq(t, `{"messages":["11","12","13"]}`, string(p.req.Body))
		})
	}
}

func TestBlockDefinitionErrors(t *testing.T) {
	tests := map[string]struct {
		nodeType FlowNodeType
		data     string
		err      string
	}{
		"too few IDs": {
			nodeType: "action_message_bulk_delete",
			data:     `{"message_ids": "11"}`,
			err:      "invalid value for message_ids: must be at least 2 IDs",
		},
		"missing required field": {
			nodeType: "action_message_bulk_delete",
			data:     `{}`,
			err:      "message_ids is required",
		},
		"not an ID": {
			nodeType: "action_message_list",
			data:     `{"channel_target": "1/../../guilds/2"}`,
			err:      "invalid value for channel_target: must be an ID",
		},
		"not a number": {
			nodeType: "action_invite_create",
			data:     `{"max_age": "one day"}`,
			err:      "invalid value for max_age: must be a whole number",
		},
		"out of range": {
			nodeType: "action_message_list",
			data:     `{"limit": 500}`,
			err:      "invalid value for limit: must be at most 100",
		},
		"not a boolean": {
			nodeType: "action_invite_create",
			data:     `{"temporary": "yes"}`,
			err:      "invalid value for temporary: must be true or false",
		},
		"too long": {
			nodeType: "action_role_create",
			data:     `{"name": "` + strings.Repeat("a", 101) + `"}`,
			err:      "invalid value for name: must be at most 100 characters",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := executeBlock(t, &blockTestProvider{}, tt.nodeType, tt.data)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestBlockDefinitionCredits(t *testing.T) {
	node := &CompiledFlowNode{Type: "action_role_create"}
	assert.Equal(t, 1, node.CreditsCost())
}

// CreditsCost only reads the definition's credits for request blocks.
func TestRequestBlocksHaveCredits(t *testing.T) {
	for _, def := range blockDefinitions {
		if def.Run.Kind == "request" {
			assert.NotNil(t, def.Credits, def.Type)
		}
	}
}

// Fields named like a setting of FlowNodeData read it, which only works for
// text settings like channel_target.
func TestBlockDefinitionFieldSettings(t *testing.T) {
	dataType := reflect.TypeOf(FlowNodeData{})
	for _, block := range blockDefinitions {
		for _, field := range block.Fields {
			if i, ok := flowNodeDataFields[field.Name]; ok {
				assert.Equalf(t, reflect.String, dataType.Field(i).Type.Kind(), "%s.%s", block.Type, field.Name)
			}
		}
	}
}

// A placeholder that turns out empty doesn't fall back to the server the flow
// runs in, which could ban someone on the wrong server.
func TestBlockDefinitionNoFallbackForEmptyPlaceholder(t *testing.T) {
	p := &blockTestProvider{response: `{"id":"7"}`}
	_, err := executeBlock(t, p, "action_role_create", `{"guild_target": "{{''}}"}`)
	assert.ErrorContains(t, err, "guild_target is required")
	assert.Empty(t, p.req.Path)
}
