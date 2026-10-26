package flow

import (
	"context"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferencedParameters(t *testing.T) {
	data := FlowData{Nodes: []FlowNode{
		{ID: "a", Data: FlowNodeData{ChannelTarget: "{{params.log_channel}}"}},
		{ID: "b", Data: FlowNodeData{AuditLogReason: "Be {{ params.tone }} in <#{{params.log_channel}}>, not params.ignored"}},
		{ID: "c", Data: FlowNodeData{ChannelTarget: "{{webhook.data.params.other}}"}},
		{ID: "d", Data: FlowNodeData{ChannelTarget: "{{1<params.max && params.min>0}}"}},
		{ID: "e", Type: FlowNodeTypeNote, Data: FlowNodeData{Fields: map[string]any{"note_text": "Set {{params.noted}}"}}},
	}}
	assert.Equal(t, []string{"log_channel", "tone", "max", "min"}, data.ReferencedParameters())
}

func TestParametersInTemplates(t *testing.T) {
	data := FlowData{
		Nodes:      []FlowNode{{ID: "entry", Type: FlowNodeTypeEntryCommand}},
		Parameters: []FlowParameter{{Key: "greeting", Type: FlowParameterTypeText, Value: "Hello"}},
	}
	node, err := CompileCommand(data)
	require.NoError(t, err)

	evalCtx := eval.NewContext(eval.Env{"params": node.Parameters})
	res, err := eval.EvalTemplateToString(context.Background(), "{{params.greeting}} and {{params.missing}}!", evalCtx)
	require.NoError(t, err)
	assert.Equal(t, "Hello and !", res)
}

func TestUsedParameters(t *testing.T) {
	data := FlowData{
		Nodes: []FlowNode{{ID: "a", Data: FlowNodeData{ChannelTarget: "{{params.log}} {{params.new}}"}}},
		Parameters: []FlowParameter{
			{Key: "log", Label: "Log", Type: FlowParameterTypeChannel, Value: "1"},
			{Key: "unused", Type: FlowParameterTypeText},
		},
	}
	assert.Equal(t, []FlowParameter{
		{Key: "log", Label: "Log", Type: FlowParameterTypeChannel, Value: "1"},
		{Key: "new", Type: FlowParameterTypeText, Required: true},
	}, data.UsedParameters())
}

func TestWithParameterValues(t *testing.T) {
	data := FlowData{Parameters: []FlowParameter{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}}}

	updated, changed := data.WithParameterValues(map[string]string{"a": "new", "c": "x"})
	assert.True(t, changed)
	assert.Equal(t, map[string]string{"a": "new", "b": "2"}, updated.ParameterValues())
	// The flow it was called on keeps its values.
	assert.Equal(t, "1", data.Parameters[0].Value)

	_, changed = data.WithParameterValues(map[string]string{"a": "1"})
	assert.False(t, changed)
}

func TestParameterCheckValue(t *testing.T) {
	channel := FlowParameter{Key: "log", Type: FlowParameterTypeChannel}
	assert.NoError(t, channel.CheckValue(""))
	assert.NoError(t, channel.CheckValue("123456789012345678"))
	assert.EqualError(t, channel.CheckValue("general"), "log must be the ID of a channel")
	assert.NoError(t, FlowParameter{Key: "x", Type: FlowParameterTypeText}.CheckValue("anything"))
	// Saving a flow checks the values too.
	assert.Error(t, FlowData{Parameters: []FlowParameter{{Key: "log", Type: FlowParameterTypeChannel, Value: "general"}}}.Validate())
}
