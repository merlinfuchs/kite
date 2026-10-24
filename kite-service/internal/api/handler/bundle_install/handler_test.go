package bundleinstall

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemapMessageTemplates(t *testing.T) {
	flowSource := flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "entry", Type: flow.FlowNodeTypeEntryCommand},
		{ID: "send", Type: flow.FlowNodeTypeActionMessageCreate, Data: flow.FlowNodeData{MessageTemplateID: "panel"}},
	}}

	require.NoError(t, remapMessageTemplates(&flowSource, map[string]string{"panel": "msg_1"}))
	assert.Equal(t, "", flowSource.Nodes[0].Data.MessageTemplateID)
	assert.Equal(t, "msg_1", flowSource.Nodes[1].Data.MessageTemplateID)

	other := flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "send", Type: flow.FlowNodeTypeActionMessageCreate, Data: flow.FlowNodeData{MessageTemplateID: "elsewhere"}},
	}}
	assert.Error(t, remapMessageTemplates(&other, map[string]string{"panel": "msg_1"}))
}
