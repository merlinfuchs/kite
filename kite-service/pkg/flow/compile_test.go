package flow

import (
	"encoding/json"
	"strings"
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

func commandArgumentsFlow(order []string) FlowData {
	argument := func(id string, required bool) FlowNode {
		return FlowNode{
			ID:   id,
			Type: FlowNodeTypeOptionCommandArgument,
			Data: FlowNodeData{
				Name:                    id,
				Description:             id,
				CommandArgumentType:     CommandArgumentTypeString,
				CommandArgumentRequired: required,
			},
		}
	}

	return FlowData{
		Nodes: []FlowNode{
			{
				ID:   "entry",
				Type: FlowNodeTypeEntryCommand,
				Data: FlowNodeData{
					Name:                 "ban",
					Description:          "Ban a user",
					CommandArgumentOrder: order,
				},
			},
			argument("reason", false),
			argument("user", true),
			argument("days", false),
			argument("duration", true),
		},
		Edges: []FlowEdge{
			{Source: "reason", Target: "entry"},
			{Source: "user", Target: "entry"},
			{Source: "days", Target: "entry"},
			{Source: "duration", Target: "entry"},
		},
	}
}

func TestFlowCommandArgumentOrder(t *testing.T) {
	tests := []struct {
		name     string
		order    []string
		expected []string
	}{
		{
			name:     "connection order without an order",
			expected: []string{"duration", "user", "reason", "days"},
		},
		{
			name:     "follows the order",
			order:    []string{"user", "duration", "days", "reason"},
			expected: []string{"user", "duration", "days", "reason"},
		},
		{
			name:     "required arguments stay first",
			order:    []string{"days", "user", "reason", "duration"},
			expected: []string{"user", "duration", "days", "reason"},
		},
		{
			name:     "missing arguments come last",
			order:    []string{"days", "user", "deleted"},
			expected: []string{"user", "duration", "days", "reason"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry, err := CompileCommand(commandArgumentsFlow(test.order))
			require.NoError(t, err)

			names := []string{}
			for _, o := range entry.CommandArguments() {
				names = append(names, o.Name())
			}
			assert.Equal(t, test.expected, names)
		})
	}
}

func webhookFlow(action FlowNodeType) FlowData {
	return FlowData{
		Nodes: []FlowNode{
			{
				ID:   "0",
				Type: FlowNodeTypeEntryEvent,
				Data: FlowNodeData{
					EventType:   EventTypeWebhook,
					Description: "On webhook request",
				},
			},
			{ID: "1", Type: action},
		},
		Edges: []FlowEdge{{Source: "0", Target: "1"}},
	}
}

func TestFlowCompileWebhook(t *testing.T) {
	got, err := CompileEventListener(webhookFlow(FlowNodeTypeActionMessageCreate))
	require.NoError(t, err)
	assert.True(t, got.IsWebhookEntry())
	assert.False(t, got.IsScheduleEntry())
	assert.Equal(t, EventTypeWebhook, got.EventListenerType())

	for _, action := range []FlowNodeType{
		FlowNodeTypeActionResponseCreate,
		FlowNodeTypeActionResponseDefer,
		FlowNodeTypeSuspendResponseModal,
	} {
		_, err := CompileEventListener(webhookFlow(action))
		assert.Error(t, err, "%s should be rejected in webhook flows", action)
	}
}

func TestFlowCompileWebhookAllowsResponsesInButtonBranches(t *testing.T) {
	data := webhookFlow(FlowNodeTypeActionMessageCreate)
	data.Nodes = append(data.Nodes, FlowNode{ID: "2", Type: FlowNodeTypeActionResponseCreate})
	data.Edges = append(data.Edges, FlowEdge{Source: "1", Target: "2", SourceHandle: null.StringFrom("component_1")})

	_, err := CompileEventListener(data)
	assert.NoError(t, err, "a clicked button has an interaction to respond to")
}

func TestCheckInteractionBlocksInDiscordEventListeners(t *testing.T) {
	data := FlowData{
		Nodes: []FlowNode{
			{ID: "0", Type: FlowNodeTypeEntryEvent, Data: FlowNodeData{EventType: "message_create", Description: "On message"}},
			{ID: "1", Type: FlowNodeTypeActionMessageCreate},
			{ID: "2", Type: FlowNodeTypeActionResponseCreate},
		},
		Edges: []FlowEdge{
			{Source: "0", Target: "1"},
			{Source: "1", Target: "2", SourceHandle: null.StringFrom("component_1")},
		},
	}

	entry, err := CompileEventListener(data)
	require.NoError(t, err)
	assert.NoError(t, entry.CheckInteractionBlocks(), "a clicked button has an interaction to respond to")

	// Listeners saved before the check still compile and run.
	data.Edges[1].SourceHandle = null.String{}
	entry, err = CompileEventListener(data)
	require.NoError(t, err)
	assert.ErrorContains(t, entry.CheckInteractionBlocks(), "action_response_create")
}

func TestCommandValidationMatchesDiscordLimits(t *testing.T) {
	entry := FlowNodeData{Name: "ping", Description: "Ping"}
	assert.NoError(t, entry.Validate(FlowNodeTypeEntryCommand))

	entry.Description = "  "
	assert.Error(t, entry.Validate(FlowNodeTypeEntryCommand), "blank description")

	arg := FlowNodeData{Name: "text", Description: "Text", CommandArgumentType: CommandArgumentTypeString}
	assert.NoError(t, arg.Validate(FlowNodeTypeOptionCommandArgument))

	arg.Description = " "
	assert.Error(t, arg.Validate(FlowNodeTypeOptionCommandArgument), "blank description")
	arg.Description = "Text"

	arg.CommandArgumentMaxLength = 6001
	assert.Error(t, arg.Validate(FlowNodeTypeOptionCommandArgument), "max length over 6000")
	arg.CommandArgumentMaxLength = 6000
	assert.NoError(t, arg.Validate(FlowNodeTypeOptionCommandArgument))

	arg.CommandArgumentMaxValue = 1e16
	assert.Error(t, arg.Validate(FlowNodeTypeOptionCommandArgument), "max value over 2^53")
}

func TestCheckCommandArgumentsRejectsDuplicateNames(t *testing.T) {
	data := FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "ping", Description: "Ping"}},
			{ID: "a", Type: FlowNodeTypeOptionCommandArgument, Data: FlowNodeData{Name: "user", Description: "User", CommandArgumentType: CommandArgumentTypeUser}},
			{ID: "b", Type: FlowNodeTypeOptionCommandArgument, Data: FlowNodeData{Name: "user", Description: "Other", CommandArgumentType: CommandArgumentTypeString}},
		},
		Edges: []FlowEdge{{Source: "a", Target: "entry"}, {Source: "b", Target: "entry"}},
	}

	// Commands saved before the check still compile and run.
	entry, err := CompileCommand(data)
	require.NoError(t, err)
	assert.ErrorContains(t, entry.CheckCommandArguments(), "more than one argument is named user")

	data.Nodes[2].Data.Name = "text"
	entry, err = CompileCommand(data)
	require.NoError(t, err)
	assert.NoError(t, entry.CheckCommandArguments())
}

// Notes are text the editor puts on the canvas next to the blocks. They're
// saved with the flow, but nothing connects to them, so they never run.
func TestFlowNotesAreKeptAndNeverRun(t *testing.T) {
	raw := `{
		"nodes": [
			{"id": "0", "type": "entry_command", "data": {"name": "ping", "description": "Pong!"}, "position": {"x": 0, "y": 0}},
			{"id": "1", "type": "action_response_create", "data": {"message_data": {"content": "Pong!"}}, "position": {"x": 0, "y": 200}},
			{"id": "note_abc", "type": "note", "data": {"note_title": "Ping", "note_text": "Replies with **Pong!**", "note_color": "blue"}, "position": {"x": 400, "y": 0}, "width": 300, "height": 120}
		],
		"edges": [{"id": "1", "source": "0", "target": "1"}]
	}`

	var data FlowData
	require.NoError(t, json.Unmarshal([]byte(raw), &data))
	require.NoError(t, data.Validate())

	saved, err := json.Marshal(data)
	require.NoError(t, err)
	var again FlowData
	require.NoError(t, json.Unmarshal(saved, &again))
	note := again.Nodes[2]
	assert.Equal(t, "Replies with **Pong!**", note.Data.Setting("note_text"))
	assert.Equal(t, "blue", note.Data.Setting("note_color"))
	assert.Equal(t, 300.0, note.Width)

	// Resizing a note isn't an edit of what the flow does.
	hash, err := again.ContentHash()
	require.NoError(t, err)
	again.Nodes[2].Width = 500
	resized, err := again.ContentHash()
	require.NoError(t, err)
	assert.Equal(t, hash, resized)

	entry, err := CompileCommand(again)
	require.NoError(t, err)
	require.Len(t, entry.Children.Default, 1)
	assert.Equal(t, FlowNodeTypeActionResponseCreate, entry.Children.Default[0].Type)
	assert.Nil(t, firstInExecution([]*CompiledFlowNode{entry}, func(n *CompiledFlowNode) bool {
		return n.Type == FlowNodeTypeNote
	}))
}

func TestFlowNoteValidation(t *testing.T) {
	note := FlowNode{ID: "note_abc", Type: FlowNodeTypeNote, Data: FlowNodeData{Fields: map[string]any{"note_text": strings.Repeat("a", 2000)}}}
	data := FlowData{Nodes: []FlowNode{note}}
	assert.NoError(t, data.Validate())

	note.Data.Fields["note_text"] = strings.Repeat("a", 2001)
	assert.Error(t, FlowData{Nodes: []FlowNode{note}}.Validate())

	note.Data.Fields["note_text"] = "fine"
	block := FlowNode{ID: "1", Type: FlowNodeTypeActionResponseCreate, Data: FlowNodeData{}}
	data = FlowData{
		Nodes: []FlowNode{block, note},
		Edges: []FlowEdge{{ID: "1", Source: "1", Target: "note_abc"}},
	}
	assert.Error(t, data.Validate())
}
