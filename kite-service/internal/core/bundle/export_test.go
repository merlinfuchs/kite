package bundle

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport(t *testing.T) {
	const (
		channel = "100000000000000001"
		role    = "100000000000000002"
		user    = "100000000000000003"
	)
	targets := map[string]Target{
		channel: {Type: SettingTypeChannel, Name: "General Chat"},
		role:    {Type: SettingTypeRole, Name: "Mods"},
	}

	cmd := &model.Command{ID: "cmd1", FlowSource: flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "entry", Type: flow.FlowNodeTypeEntryCommand, Data: flow.FlowNodeData{Name: "ping"}},
		{ID: "send", Type: flow.FlowNodeTypeActionMessageCreate, Data: flow.FlowNodeData{
			ChannelTarget:     channel,
			MessageTemplateID: "msg1",
		}},
		{ID: "role", Type: flow.FlowNodeTypeActionMemberRoleAdd, Data: flow.FlowNodeData{RoleTarget: role, UserTarget: user}},
		{ID: "count", Type: flow.FlowNodeTypeActionVariableSet, Data: flow.FlowNodeData{VariableID: "var1"}},
	}}}
	msg := &model.Message{ID: "msg1", Name: "Panel", Data: message.MessageData{
		Content: "Ask in <#" + channel + ">, <@&" + role + "> will help.",
	}}
	variable := &model.Variable{ID: "var1", Name: "count", Scoped: true}

	b, unmatched, err := Export("My App", ExportItems{
		Commands:  []*model.Command{cmd},
		Messages:  []*model.Message{msg},
		Variables: []*model.Variable{variable},
	}, func(id string) (Target, bool) {
		target, ok := targets[id]
		return target, ok
	})
	require.NoError(t, err)

	assert.Equal(t, []string{user}, unmatched)
	assert.Equal(t, []Setting{
		{Key: "channel_general_chat", Label: "#General Chat", Description: "The channel to use where the exported app used #General Chat.", Type: SettingTypeChannel, Required: true},
		{Key: "role_mods", Label: "@Mods", Description: "The role to use where the exported app used @Mods.", Type: SettingTypeRole, Required: true},
	}, b.Settings)

	raw, err := json.Marshal(b)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), channel)
	assert.NotContains(t, string(raw), role)
	assert.Contains(t, string(raw), user)

	// It imports like any bundle, with the importing app's channel and role.
	parsed, err := Parse(raw)
	require.NoError(t, err)
	sel, err := parsed.Select([]string{"cmd1", "msg1", "var1"}, map[string]string{
		"channel_general_chat": "200000000000000001",
		"role_mods":            "200000000000000002",
	})
	require.NoError(t, err)
	assert.Equal(t, "200000000000000001", sel.Commands[0].FlowSource.Nodes[1].Data.ChannelTarget)
	assert.Equal(t, "200000000000000002", sel.Commands[0].FlowSource.Nodes[2].Data.RoleTarget)
	assert.Equal(t, "msg1", sel.Commands[0].FlowSource.Nodes[1].Data.MessageTemplateID)
	assert.Equal(t, "Ask in <#200000000000000001>, <@&200000000000000002> will help.", sel.Messages[0].Data.Content)
	assert.Equal(t, []Variable{{ID: "var1", Name: "count", Scoped: true}}, sel.Variables)
}

func TestSettingKey(t *testing.T) {
	settings := []Setting{{Key: "channel_general"}}
	assert.Equal(t, "channel_general_2", settingKey(Target{Type: SettingTypeChannel, Name: "general"}, settings))
	assert.Equal(t, "role", settingKey(Target{Type: SettingTypeRole, Name: "✨"}, nil))
	assert.True(t, strings.HasPrefix(settingKey(Target{Type: SettingTypeRole, Name: "Mod Team!"}, nil), "role_mod_team"))
}
