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

// variableScope evaluates the scope of a variable block. The variable decides
// whether there is one: blocks keep their scope when it's switched to
// unscoped, and a scoped variable has no value without a scope.
func variableScope(n *CompiledFlowNode, ctx *FlowContext) (null.String, error) {
	scoped, err := ctx.Variable.VariableScoped(ctx, n.Data.VariableID)
	if err != nil || !scoped {
		return null.String{}, err
	}

	scope, err := ctx.EvalTemplate(n.Data.VariableScope)
	if err != nil {
		return null.String{}, err
	}
	if s := scope.String(); s != "" {
		return null.StringFrom(s), nil
	}
	return null.String{}, errors.New("the variable is scoped, but the scope is empty")
}

func executeActionVariableSet(n *CompiledFlowNode, ctx *FlowContext) error {
	scope, err := variableScope(n, ctx)
	if err != nil {
		return err
	}

	evalValue := ctx.EvalTemplate
	switch n.Data.VariableOperation {
	case provider.VariableOperationAppend, provider.VariableOperationPrepend:
		// Spaces between the joined texts are part of the value.
		evalValue = ctx.EvalTemplateKeepSpace
	}

	value, err := evalValue(n.Data.VariableValue)
	if err != nil {
		return err
	}

	newValue, err := ctx.Variable.UpdateVariable(
		ctx,
		n.Data.VariableID,
		scope,
		n.Data.VariableOperation,
		value,
	)
	if err != nil {
		return err
	}

	ctx.StoreNodeResult(n, newValue)
	return n.ExecuteChildren(ctx)
}

func executeActionVariableDelete(n *CompiledFlowNode, ctx *FlowContext) error {
	scope, err := variableScope(n, ctx)
	if err != nil {
		return err
	}

	err = ctx.Variable.DeleteVariable(
		ctx,
		n.Data.VariableID,
		scope,
	)
	if err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}

func executeActionVariableGet(n *CompiledFlowNode, ctx *FlowContext) error {
	scope, err := variableScope(n, ctx)
	if err != nil {
		return err
	}

	val, err := ctx.Variable.Variable(
		ctx,
		n.Data.VariableID,
		scope,
	)
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return err
	}

	ctx.StoreNodeResult(n, val)
	return n.ExecuteChildren(ctx)
}
