package bundleinstall

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
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

func TestMessageContentHash(t *testing.T) {
	button := flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "entry", Type: flow.FlowNodeTypeEntryComponentButton, Position: flow.FlowNodePosition{X: 1}},
	}}
	hash := func(data message.MessageData, flows map[string]flow.FlowData) string {
		h, err := messageContentHash(data, flows)
		require.NoError(t, err)
		return h
	}

	want := hash(message.MessageData{Content: "Verify"}, map[string]flow.FlowData{"verify": button})

	moved := button
	moved.Nodes = []flow.FlowNode{{ID: "entry", Type: flow.FlowNodeTypeEntryComponentButton, Position: flow.FlowNodePosition{X: 50}}}
	assert.Equal(t, want, hash(message.MessageData{Content: "Verify"}, map[string]flow.FlowData{"verify": moved}))
	assert.NotEqual(t, want, hash(message.MessageData{Content: "Verify now"}, map[string]flow.FlowData{"verify": button}))
	assert.NotEqual(t, want, hash(message.MessageData{Content: "Verify"}, nil))
}
