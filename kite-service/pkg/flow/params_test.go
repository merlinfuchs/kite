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
	}}
	assert.Equal(t, []string{"log_channel", "tone"}, data.ReferencedParameters())
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
