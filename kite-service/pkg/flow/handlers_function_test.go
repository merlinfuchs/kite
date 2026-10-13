package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseFlowData(t *testing.T, raw string) FlowData {
	t.Helper()
	var data FlowData
	require.NoError(t, json.Unmarshal([]byte(raw), &data))
	return data
}

// callerFlow is an event listener that calls the function f and stores what
// it returns in the temporary variable called.
const callerFlow = `{
	"nodes": [
		{"id": "entry", "type": "entry_event", "data": {"event_type": "message_create", "description": "Test"}},
		{"id": "call", "type": "action_function_call", "data": {
			"function_id": "f",
			"temporary_name": "called",
			"function_arguments": [
				{"name": "a", "value": "{{1 + 2}}"},
				{"name": "b", "value": "x"}
			]
		}}
	],
	"edges": [
		{"id": "e1", "source": "entry", "target": "call"}
	]
}`

func runCaller(t *testing.T, functions map[string]string) (*FlowContext, error) {
	t.Helper()

	flows := make(map[string]FlowData, len(functions))
	for id, raw := range functions {
		flows[id] = parseFlowData(t, raw)
	}

	entry, err := CompileEventListener(parseFlowData(t, callerFlow))
	require.NoError(t, err)

	c := NewContext(
		context.Background(),
		5*time.Second,
		&TestContextData{},
		FlowProviders{
			Discord:  &TestDiscordProvider{},
			Log:      &provider.MockLogProvider{},
			Function: &MockFunctionProvider{Functions: flows},
		},
		FlowContextLimits{
			MaxStackDepth: 100,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	// Set before the call, the function mustn't see it.
	c.SetTemporary("secret", thing.NewString("caller"))

	return c, entry.Execute(c)
}

func TestFunctionCallReturnsValue(t *testing.T) {
	c, err := runCaller(t, map[string]string{"f": `{
		"nodes": [
			{"id": "fentry", "type": "entry_function", "data": {
				"name": "Double",
				"function_parameters": [{"name": "a"}, {"name": "b"}]
			}},
			{"id": "double", "type": "action_expression_evaluate", "data": {"expression": "arg('a') * 2", "temporary_name": "double"}},
			{"id": "ret", "type": "control_function_return", "data": {"function_return_value": "{{var('double')}}-{{arg('b')}}-{{var('secret')}}"}}
		],
		"edges": [
			{"id": "e1", "source": "fentry", "target": "double"},
			{"id": "e2", "source": "double", "target": "ret"}
		]
	}`})
	require.NoError(t, err)

	assert.Equal(t, "6-x-", c.GetNodeResult("call").String())
	assert.Equal(t, "6-x-", c.GetTemporary("called").String())
	// The function's results and temporary variables stay in the function.
	assert.True(t, c.GetNodeResult("double").IsNil())
	assert.True(t, c.GetTemporary("double").IsNil())
	// Only the blocks of the function cost credits, not the call.
	assert.Equal(t, 1, c.CreditsUsed())
}

func TestFunctionStopsAtReturn(t *testing.T) {
	c, err := runCaller(t, map[string]string{"f": `{
		"nodes": [
			{"id": "fentry", "type": "entry_function", "data": {"name": "First"}},
			{"id": "ret1", "type": "control_function_return", "data": {"function_return_value": "first"}},
			{"id": "ret2", "type": "control_function_return", "data": {"function_return_value": "second"}}
		],
		"edges": [
			{"id": "e1", "source": "fentry", "target": "ret1"},
			{"id": "e2", "source": "fentry", "target": "ret2"}
		]
	}`})
	require.NoError(t, err)
	assert.Equal(t, "first", c.GetNodeResult("call").String())
}

func TestFunctionWithoutReturn(t *testing.T) {
	c, err := runCaller(t, map[string]string{"f": `{
		"nodes": [{"id": "fentry", "type": "entry_function", "data": {"name": "Nothing"}}],
		"edges": []
	}`})
	require.NoError(t, err)
	assert.True(t, c.GetNodeResult("call").IsNil())
}

func TestFunctionRecursionIsLimited(t *testing.T) {
	_, err := runCaller(t, map[string]string{"f": `{
		"nodes": [
			{"id": "fentry", "type": "entry_function", "data": {"name": "Forever"}},
			{"id": "again", "type": "action_function_call", "data": {"function_id": "f"}}
		],
		"edges": [{"id": "e1", "source": "fentry", "target": "again"}]
	}`})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "levels deep")
}

func TestFunctionNotFound(t *testing.T) {
	_, err := runCaller(t, map[string]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "function not found")
}

func TestReturnOutsideFunction(t *testing.T) {
	entry, err := CompileEventListener(parseFlowData(t, `{
		"nodes": [
			{"id": "entry", "type": "entry_event", "data": {"event_type": "message_create", "description": "Test"}},
			{"id": "ret", "type": "control_function_return", "data": {"function_return_value": "x"}}
		],
		"edges": [{"id": "e1", "source": "entry", "target": "ret"}]
	}`))
	require.NoError(t, err)

	c := NewContext(context.Background(), time.Second, &TestContextData{}, FlowProviders{}, FlowContextLimits{}, eval.NewContext(eval.Env{}), nil)
	defer c.Cancel()

	err = entry.Execute(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only be used in functions")
}

func TestCompileFunction(t *testing.T) {
	entry, err := CompileFunction(parseFlowData(t, `{
		"nodes": [{"id": "fentry", "type": "entry_function", "data": {"name": "Greet", "description": "Says hi"}}],
		"edges": []
	}`))
	require.NoError(t, err)
	assert.Equal(t, "Greet", entry.FunctionName())
	assert.Equal(t, "Says hi", entry.FunctionDescription())

	_, err = CompileFunction(parseFlowData(t, `{
		"nodes": [{"id": "entry", "type": "entry_command", "data": {"name": "test", "description": "Test"}}],
		"edges": []
	}`))
	assert.Error(t, err, "a flow without a function entry isn't a function")

	_, err = CompileFunction(parseFlowData(t, `{
		"nodes": [
			{"id": "fentry", "type": "entry_function", "data": {"name": "Ask"}},
			{"id": "modal", "type": "suspend_response_modal", "data": {}}
		],
		"edges": [{"id": "e1", "source": "fentry", "target": "modal"}]
	}`))
	assert.ErrorContains(t, err, "can't be used in functions")

	_, err = CompileFunction(parseFlowData(t, `{
		"nodes": [
			{"id": "fentry", "type": "entry_function", "data": {"name": "Buttons"}},
			{"id": "msg", "type": "action_message_create", "data": {}},
			{"id": "log", "type": "action_log", "data": {}}
		],
		"edges": [
			{"id": "e1", "source": "fentry", "target": "msg"},
			{"id": "e2", "source": "msg", "target": "log", "sourceHandle": "component_1"}
		]
	}`))
	assert.ErrorContains(t, err, "can't be used in functions")
}

func TestFunctionEntryValidation(t *testing.T) {
	valid := FlowNodeData{
		Name:               "Greet",
		FunctionParameters: []FunctionParameterData{{Name: "user_id"}},
	}
	assert.NoError(t, valid.Validate(FlowNodeTypeEntryFunction))

	assert.Error(t, FlowNodeData{}.Validate(FlowNodeTypeEntryFunction), "name is required")

	invalidName := valid
	invalidName.FunctionParameters = []FunctionParameterData{{Name: "User ID"}}
	assert.Error(t, invalidName.Validate(FlowNodeTypeEntryFunction))

	duplicate := valid
	duplicate.FunctionParameters = []FunctionParameterData{{Name: "a"}, {Name: "a"}}
	assert.Error(t, duplicate.Validate(FlowNodeTypeEntryFunction))
}

func TestFunctionReturnStopsCaller(t *testing.T) {
	for _, stop := range []bool{false, true} {
		caller := parseFlowData(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_event", "data": {"event_type": "message_create", "description": "Test"}},
				{"id": "call", "type": "action_function_call", "data": {"function_id": "check"}},
				{"id": "after", "type": "action_expression_evaluate", "data": {"expression": "1", "temporary_name": "after"}}
			],
			"edges": [
				{"id": "e1", "source": "entry", "target": "call"},
				{"id": "e2", "source": "call", "target": "after"}
			]
		}`)

		// check calls guard, so stopping also has to end the function between
		// the guard and the caller.
		functions := map[string]FlowData{
			"check": parseFlowData(t, `{
				"nodes": [
					{"id": "centry", "type": "entry_function", "data": {"name": "Check"}},
					{"id": "inner", "type": "action_function_call", "data": {"function_id": "guard"}},
					{"id": "cret", "type": "control_function_return", "data": {"function_return_value": "check"}}
				],
				"edges": [
					{"id": "e1", "source": "centry", "target": "inner"},
					{"id": "e2", "source": "inner", "target": "cret"}
				]
			}`),
			"guard": parseFlowData(t, fmt.Sprintf(`{
				"nodes": [
					{"id": "gentry", "type": "entry_function", "data": {"name": "Guard"}},
					{"id": "gret", "type": "control_function_return", "data": {"function_return_value": "denied", "function_return_stop": %t}}
				],
				"edges": [{"id": "e1", "source": "gentry", "target": "gret"}]
			}`, stop)),
		}

		entry, err := CompileEventListener(caller)
		require.NoError(t, err)

		c := NewContext(
			context.Background(),
			5*time.Second,
			&TestContextData{},
			FlowProviders{
				Discord:  &TestDiscordProvider{},
				Log:      &provider.MockLogProvider{},
				Function: &MockFunctionProvider{Functions: functions},
			},
			FlowContextLimits{MaxStackDepth: 100, MaxOperations: 1000, MaxCredits: 1000},
			eval.NewContext(eval.Env{}),
			nil,
		)
		require.NoError(t, entry.Execute(c))
		c.Cancel()

		if stop {
			assert.True(t, c.GetTemporary("after").IsNil(), "blocks after the call mustn't run")
			// check stopped before its own Return block, so it returned nothing.
			assert.True(t, c.GetNodeResult("call").IsNil())
		} else {
			assert.False(t, c.GetTemporary("after").IsNil())
			assert.Equal(t, "check", c.GetNodeResult("call").String())
		}
	}
}
