package bundleinstall

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleIDsRemap(t *testing.T) {
	ids := bundleIDs{
		messages:  map[string]string{"panel": "msg_1"},
		variables: map[string]string{"count": "var_1"},
	}

	flowSource := flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "entry", Type: flow.FlowNodeTypeEntryCommand},
		{ID: "send", Type: flow.FlowNodeTypeActionMessageCreate, Data: flow.FlowNodeData{MessageTemplateID: "panel"}},
		{ID: "set", Type: flow.FlowNodeTypeActionVariableSet, Data: flow.FlowNodeData{VariableID: "count"}},
	}}
	require.NoError(t, ids.remap(&flowSource))
	assert.Equal(t, "", flowSource.Nodes[0].Data.MessageTemplateID)
	assert.Equal(t, "msg_1", flowSource.Nodes[1].Data.MessageTemplateID)
	assert.Equal(t, "var_1", flowSource.Nodes[2].Data.VariableID)

	other := flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "send", Type: flow.FlowNodeTypeActionMessageCreate, Data: flow.FlowNodeData{MessageTemplateID: "elsewhere"}},
	}}
	assert.Error(t, ids.remap(&other))

	other = flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "set", Type: flow.FlowNodeTypeActionVariableSet, Data: flow.FlowNodeData{VariableID: "elsewhere"}},
	}}
	assert.Error(t, ids.remap(&other))
}
