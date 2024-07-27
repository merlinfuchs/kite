package flow

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
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
				MessageData: api.SendMessageData{
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
	}

	expected.Children = []*CompiledFlowNode{
		{
			ID:   "1",
			Type: FlowNodeTypeActionResponseCreate,
			Parents: []*CompiledFlowNode{
				expected,
			},
			Data: FlowNodeData{
				MessageData: api.SendMessageData{
					Content: "Pong!",
				},
			},
		},
	}

	got, err := CompileCommand(flowCommandInput)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
