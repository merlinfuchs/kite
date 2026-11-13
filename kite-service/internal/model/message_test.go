package model

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageContentHash(t *testing.T) {
	button := flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "entry", Type: flow.FlowNodeTypeEntryComponentButton, Position: flow.FlowNodePosition{X: 1}},
	}}
	hash := func(data message.MessageData, flows map[string]flow.FlowData) string {
		h, err := (&Message{Data: data, FlowSources: flows}).ContentHash()
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
