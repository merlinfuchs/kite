package flow

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionVariableSet:    executeActionVariableSet,
		FlowNodeTypeActionVariableDelete: executeActionVariableDelete,
		FlowNodeTypeActionVariableGet:    executeActionVariableGet,
		FlowNodeTypeActionVariableList:   executeActionVariableList,
	})
}

// The variable blocks are presets of the operations of stored variables in
// operation_handlers_variables.go, with their settings as templates.

// blockVariableCall is the variableCall of a variable block, with the
// template of its scope, or of the start of the scopes for a list.
func blockVariableCall(n *CompiledFlowNode, ctx *FlowContext, scope string) variableCall {
	return variableCall{
		ctx: ctx,
		id:  n.Data.VariableID,
		scope: func() (string, error) {
			v, err := ctx.EvalTemplate(scope)
			return v.String(), err
		},
	}
}

func executeActionVariableSet(n *CompiledFlowNode, ctx *FlowContext) error {
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

	newValue, err := updateVariable(blockVariableCall(n, ctx, n.Data.VariableScope), n.Data.VariableOperation, value)
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, newValue)
	return n.ExecuteChildren(ctx)
}

func executeActionVariableDelete(n *CompiledFlowNode, ctx *FlowContext) error {
	if err := deleteVariable(blockVariableCall(n, ctx, n.Data.VariableScope)); err != nil {
		return err
	}
	return n.ExecuteChildren(ctx)
}

func executeActionVariableGet(n *CompiledFlowNode, ctx *FlowContext) error {
	value, err := getVariable(blockVariableCall(n, ctx, n.Data.VariableScope))
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, value)
	return n.ExecuteChildren(ctx)
}

func executeActionVariableList(n *CompiledFlowNode, ctx *FlowContext) error {
	limit, err := variableListNumber(ctx, n.Data.VariableLimit, defaultVariableListLimit, 1, maxVariableListLimit)
	if err != nil {
		return fmt.Errorf("the number of values must be a whole number from 1 to %d", maxVariableListLimit)
	}
	skip, err := variableListNumber(ctx, n.Data.VariableSkip, 0, 0, maxVariableListSkip)
	if err != nil {
		return fmt.Errorf("the number of values to skip must be a whole number from 0 to %d", maxVariableListSkip)
	}

	values, err := listVariable(blockVariableCall(n, ctx, n.Data.VariableScopePrefix), variableListOptions{
		Sort:  n.Data.VariableSort,
		Limit: limit,
		Skip:  skip,
	})
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, values)
	return n.ExecuteChildren(ctx)
}

// variableListNumber evaluates a number setting of the List stored variable
// values block, which is def when empty and an error outside min to max.
func variableListNumber(ctx *FlowContext, setting string, def, min, max int) (int, error) {
	if strings.TrimSpace(setting) == "" {
		return def, nil
	}
	value, err := ctx.EvalTemplate(setting)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(value.String()))
	if err != nil || n < min || n > max {
		return 0, errors.New("out of range")
	}
	return n, nil
}
