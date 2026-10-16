package flow

import (
	"errors"
	"fmt"
	"maps"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeEntryFunction:         executeEntryFunction,
		FlowNodeTypeActionFunctionCall:    executeActionFunctionCall,
		FlowNodeTypeControlFunctionReturn: executeControlFunctionReturn,
	})
}

// maxFunctionDepth is how deeply functions can call other functions,
// including themselves. It stops endless recursion with a clear error before
// the stack depth limit would.
const maxFunctionDepth = 10

// functionCall is a running call of a function.
type functionCall struct {
	depth int
	args  map[string]thing.Thing

	returned bool
	value    thing.Thing
}

// flowStop is set when a Return block stops the flow that called its
// function.
type flowStop struct {
	stopped bool
}

// hasReturned reports whether nothing else should run: the function the flow
// runs has returned, or a function stopped the whole flow.
func (c *FlowContext) hasReturned() bool {
	if c.stop != nil && c.stop.stopped {
		return true
	}
	return c.function != nil && c.function.returned
}

func executeEntryFunction(n *CompiledFlowNode, ctx *FlowContext) error {
	if ctx.function == nil {
		return traceError(n, errors.New("functions only run when a Call function block calls them"))
	}

	return n.ExecuteChildren(ctx)
}

func executeActionFunctionCall(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Data.FunctionID == "" {
		return traceError(n, errors.New("no function selected"))
	}
	if ctx.Function == nil {
		return traceError(n, errors.New("functions aren't available here"))
	}

	depth := 1
	if ctx.function != nil {
		depth = ctx.function.depth + 1
	}
	if depth > maxFunctionDepth {
		return traceError(n, fmt.Errorf("functions can only call each other %d levels deep", maxFunctionDepth))
	}

	data, err := ctx.Function.FunctionFlow(ctx, n.Data.FunctionID)
	if err != nil {
		return traceError(n, err)
	}

	entry, err := CompileFunction(data)
	if err != nil {
		return traceError(n, fmt.Errorf("invalid function: %w", err))
	}

	args := make(map[string]thing.Thing, len(n.Data.FunctionArguments))
	for _, arg := range n.Data.FunctionArguments {
		if arg.Name == "" {
			continue
		}

		value, err := ctx.EvalTemplate(arg.Value)
		if err != nil {
			return traceError(n, err)
		}
		args[arg.Name] = value
	}

	result, err := ctx.callFunction(entry, &functionCall{
		depth: depth,
		args:  args,
	})
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, result)
	return n.ExecuteChildren(ctx)
}

// callFunction runs the function with the entry block entry and returns the
// value it returned. The function runs with the interaction or event of the
// caller, so it can respond to it, but with its own block results and
// temporary variables. {{arg('name')}} gets the values passed to it. Its
// blocks count against the limits of the caller's execution.
func (c *FlowContext) callFunction(entry *CompiledFlowNode, call *functionCall) (thing.Thing, error) {
	state := NewFlowContextState()

	evalCtx := eval.Context{
		Env:      maps.Clone(c.EvalCtx.Env),
		Patchers: c.EvalCtx.Patchers,
	}
	bindStateEnv(evalCtx, state)
	evalCtx.Env["arg"] = func(name string) any {
		value, ok := call.args[name]
		if !ok {
			return nil
		}
		return eval.NewThingEnv(value)
	}

	if c.stop == nil {
		c.stop = &flowStop{}
	}

	child := &FlowContext{
		Context:           c.Context,
		FlowProviders:     c.FlowProviders,
		FlowContextLimits: c.FlowContextLimits,
		FlowContextState:  *state,
		Data:              c.Data,
		EvalCtx:           evalCtx,
		Cancel:            c.Cancel,
		function:          call,
		stop:              c.stop,
	}

	err := entry.Execute(child)

	c.operations = child.operations
	c.credits = child.credits

	if err != nil {
		return thing.Null, err
	}
	return call.value, nil
}

func executeControlFunctionReturn(n *CompiledFlowNode, ctx *FlowContext) error {
	if ctx.function == nil {
		return traceError(n, errors.New("the Return block can only be used in functions"))
	}

	value, err := ctx.EvalTemplate(n.Data.FunctionReturnValue)
	if err != nil {
		return traceError(n, err)
	}

	ctx.function.value = value
	ctx.function.returned = true

	// Like returning early from the caller as well, e.g. after a permission
	// check that already told the user they can't use the command. It stops
	// every flow and function up to the one that started the execution.
	if n.Data.FunctionReturnStop {
		ctx.stop.stopped = true
	}
	return nil
}
