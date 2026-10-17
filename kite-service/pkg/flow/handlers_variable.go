package flow

import (
	"errors"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"gopkg.in/guregu/null.v4"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionVariableSet:    executeActionVariableSet,
		FlowNodeTypeActionVariableDelete: executeActionVariableDelete,
		FlowNodeTypeActionVariableGet:    executeActionVariableGet,
	})
}

func executeActionVariableSet(n *CompiledFlowNode, ctx *FlowContext) error {
	scope, err := ctx.EvalTemplate(n.Data.VariableScope)
	if err != nil {
		return traceError(n, err)
	}

	evalValue := ctx.EvalTemplate
	switch n.Data.VariableOperation {
	case provider.VariableOperationAppend, provider.VariableOperationPrepend:
		// Spaces between the joined texts are part of the value.
		evalValue = ctx.EvalTemplateKeepSpace
	}

	value, err := evalValue(n.Data.VariableValue)
	if err != nil {
		return traceError(n, err)
	}

	newValue, err := ctx.Variable.UpdateVariable(
		ctx,
		n.Data.VariableID,
		null.NewString(scope.String(), !scope.IsEmpty()),
		n.Data.VariableOperation,
		value,
	)
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, newValue)
	return n.ExecuteChildren(ctx)
}

func executeActionVariableDelete(n *CompiledFlowNode, ctx *FlowContext) error {
	scope, err := ctx.EvalTemplate(n.Data.VariableScope)
	if err != nil {
		return traceError(n, err)
	}

	err = ctx.Variable.DeleteVariable(
		ctx,
		n.Data.VariableID,
		null.NewString(scope.String(), !scope.IsEmpty()),
	)
	if err != nil {
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}

func executeActionVariableGet(n *CompiledFlowNode, ctx *FlowContext) error {
	scope, err := ctx.EvalTemplate(n.Data.VariableScope)
	if err != nil {
		return traceError(n, err)
	}

	val, err := ctx.Variable.Variable(
		ctx,
		n.Data.VariableID,
		null.NewString(scope.String(), !scope.IsEmpty()),
	)
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, val)
	return n.ExecuteChildren(ctx)
}
