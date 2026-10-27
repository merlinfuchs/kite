package model

import (
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
)

func TestContentVersion(t *testing.T) {
	command := &Command{FlowSource: flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "entry", Type: flow.FlowNodeTypeEntryCommand},
	}}}
	version := command.ContentVersion()

	toggled := *command
	toggled.Enabled = true
	toggled.UpdatedAt = time.Now()
	assert.Equal(t, version, toggled.ContentVersion(), "not changed by enabling")

	moved := *command
	moved.FlowSource = flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "entry", Type: flow.FlowNodeTypeEntryCommand, Position: flow.FlowNodePosition{X: 10}},
	}}
	assert.NotEqual(t, version, moved.ContentVersion(), "changed by moving a block")
}
