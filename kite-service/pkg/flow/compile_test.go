package flow

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var flowCommandInput = FlowData{
	Nodes: []FlowNode{
		{
			ID:   "0",
			Type: FlowNodeTypeEntryCommand,
			Data: FlowNodeData{
				Name:        "ping",
				Description: "Pong!",
			},
		},
		{
			ID:   "1",
			Type: FlowNodeTypeActionResponseCreate,
			Data: FlowNodeData{
				MessageData: &message.MessageData{
					Content: "Pong!",
				},
			},
		},
	},
	Edges: []FlowEdge{
		{
			Source: "0",
			Target: "1",
		},
	},
}

func TestFlowCompileCommand(t *testing.T) {
	expected := &CompiledFlowNode{
		ID:   "0",
		Type: FlowNodeTypeEntryCommand,
		Data: FlowNodeData{
			Name:        "ping",
			Description: "Pong!",
		},
		Children: ConnectedFlowNodes{
			Handles: make(map[string][]*CompiledFlowNode),
		},
		Parents: ConnectedFlowNodes{
			Handles: make(map[string][]*CompiledFlowNode),
		},
	}

	expected.Children.Default = []*CompiledFlowNode{
		{
			ID:   "1",
			Type: FlowNodeTypeActionResponseCreate,
			Parents: ConnectedFlowNodes{
				Default: []*CompiledFlowNode{
					expected,
				},
				Handles: make(map[string][]*CompiledFlowNode),
			},
			Children: ConnectedFlowNodes{
				Handles: make(map[string][]*CompiledFlowNode),
			},
			Data: FlowNodeData{
				MessageData: &message.MessageData{
					Content: "Pong!",
				},
			},
		},
	}

	got, err := CompileCommand(flowCommandInput)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
