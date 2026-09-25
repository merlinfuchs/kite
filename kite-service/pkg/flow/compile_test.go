package flow

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
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

func scheduleFlow(action FlowNodeType) FlowData {
	return FlowData{
		Nodes: []FlowNode{
			{
				ID:   "0",
				Type: FlowNodeTypeEntryEvent,
				Data: FlowNodeData{
					EventType:         EventTypeScheduleCron,
					EventScheduleCron: "*/5 * * * *",
					Description:       "Every five minutes",
				},
			},
			{ID: "1", Type: action},
		},
		Edges: []FlowEdge{{Source: "0", Target: "1"}},
	}
}

func TestFlowCompileSchedule(t *testing.T) {
	got, err := CompileEventListener(scheduleFlow(FlowNodeTypeActionMessageCreate))
	require.NoError(t, err)
	assert.True(t, got.IsScheduleEntry())
	assert.Equal(t, "*/5 * * * *", got.EventScheduleCron())

	for _, action := range []FlowNodeType{
		FlowNodeTypeActionResponseCreate,
		FlowNodeTypeActionResponseDefer,
		FlowNodeTypeSuspendResponseModal,
	} {
		_, err := CompileEventListener(scheduleFlow(action))
		assert.Error(t, err, "%s should be rejected in scheduled flows", action)
	}
}

func TestFlowValidateScheduleCron(t *testing.T) {
	data := FlowNodeData{EventType: EventTypeScheduleCron, Description: "test"}
	assert.Error(t, data.Validate(FlowNodeTypeEntryEvent), "missing cron")

	data.EventScheduleCron = "not a cron"
	assert.Error(t, data.Validate(FlowNodeTypeEntryEvent), "invalid cron")

	data.EventScheduleCron = "0 * * * *"
	assert.NoError(t, data.Validate(FlowNodeTypeEntryEvent))
}

func TestFlowCompileScheduleAllowsResponsesInButtonBranches(t *testing.T) {
	data := scheduleFlow(FlowNodeTypeActionMessageCreate)
	data.Nodes = append(data.Nodes, FlowNode{ID: "2", Type: FlowNodeTypeActionResponseCreate})
	data.Edges = append(data.Edges, FlowEdge{Source: "1", Target: "2", SourceHandle: null.StringFrom("component_1")})

	_, err := CompileEventListener(data)
	assert.NoError(t, err, "a clicked button has an interaction to respond to")
}
