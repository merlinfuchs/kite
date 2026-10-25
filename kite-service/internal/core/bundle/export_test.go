package bundle

import (
	"encoding/json"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport(t *testing.T) {
	cmd := &model.Command{ID: "cmd1", FlowSource: flow.FlowData{
		Nodes: []flow.FlowNode{
			{ID: "entry", Type: flow.FlowNodeTypeEntryCommand, Data: flow.FlowNodeData{Name: "ping"}},
			{ID: "send", Type: flow.FlowNodeTypeActionMessageCreate, Data: flow.FlowNodeData{
				ChannelTarget:     "{{params.log_channel}}",
				MessageTemplateID: "msg1",
			}},
			// Used without being defined, like after the AI added the block.
			{ID: "reason", Type: flow.FlowNodeTypeActionMemberBan, Data: flow.FlowNodeData{AuditLogReason: "{{params.reason}}"}},
			{ID: "count", Type: flow.FlowNodeTypeActionVariableSet, Data: flow.FlowNodeData{VariableID: "var1"}},
		},
		Parameters: []flow.FlowParameter{
			{Key: "log_channel", Label: "Log Channel", Type: flow.FlowParameterTypeChannel, Required: true, Value: "100000000000000001"},
			// No block uses it anymore.
			{Key: "unused", Type: flow.FlowParameterTypeText, Value: "x"},
		},
	}}
	msg := &model.Message{ID: "msg1", Name: "Panel", Data: message.MessageData{Content: "Hi", Attachments: []message.MessageAttachment{{AssetID: "asset1"}}}, FlowSources: map[string]flow.FlowData{
		"button": {
			Nodes:      []flow.FlowNode{{ID: "send", Data: flow.FlowNodeData{ChannelTarget: "{{params.log_channel}}"}}},
			Parameters: []flow.FlowParameter{{Key: "log_channel", Label: "Other Label", Type: flow.FlowParameterTypeChannel, Value: "100000000000000002"}},
		},
	}}
	variable := &model.Variable{ID: "var1", Name: "count", Scoped: true}

	b, err := Export("My App", ExportItems{
		Commands:  []*model.Command{cmd},
		Messages:  []*model.Message{msg},
		Variables: []*model.Variable{variable},
	})
	require.NoError(t, err)

	// Flows with the same key share the parameter, the first one is kept.
	assert.Equal(t, []Parameter{
		{Key: "log_channel", Label: "Log Channel", Type: flow.FlowParameterTypeChannel, Required: true},
		{Key: "reason", Label: "reason", Type: flow.FlowParameterTypeText, Required: true},
	}, b.Parameters)

	// The values belong to the app it came from.
	raw, err := json.Marshal(b)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "100000000000000001")
	assert.NotContains(t, string(raw), "unused")
	assert.NotContains(t, string(raw), "asset1")

	// It imports like any bundle, with the importing app's values.
	parsed, err := Parse(raw)
	require.NoError(t, err)
	sel, err := parsed.Select([]string{"cmd1", "msg1", "var1"}, map[string]string{
		"log_channel": "200000000000000001",
		"reason":      "Spam",
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"log_channel": "200000000000000001", "reason": "Spam"}, sel.Commands[0].FlowSource.ParameterValues())
	assert.Equal(t, map[string]string{"log_channel": "200000000000000001"}, sel.Messages[0].FlowSources["button"].ParameterValues())
	assert.Equal(t, "{{params.log_channel}}", sel.Commands[0].FlowSource.Nodes[1].Data.ChannelTarget)
	assert.Equal(t, []Variable{{ID: "var1", Name: "count", Scoped: true}}, sel.Variables)
}
