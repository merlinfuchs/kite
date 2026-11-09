package flow

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
)

// maxVariableListLimit is the most values a List stored variable values block
// returns. A leaderboard or a list in a message needs far fewer.
const maxVariableListLimit = 100

const defaultVariableListLimit = 10

// maxVariableListSkip bounds how far into the values a block can page. The
// database sorts all of them either way, a far offset only adds to that.
const maxVariableListSkip = 10_000

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionVariableSet:    executeActionVariableSet,
		FlowNodeTypeActionVariableDelete: executeActionVariableDelete,
		FlowNodeTypeActionVariableGet:    executeActionVariableGet,
		FlowNodeTypeActionVariableList:   executeActionVariableList,
	})
}

var errEmptyVariableScope = errors.New("the variable is scoped, but the scope is empty")

// resolveVariableScope is the scope of a value of variable id. The variable
// decides whether there is one: blocks and code keep their scope when it's
// switched to unscoped, and a scoped variable has no value without a scope.
// scope is only called for scoped variables.
func resolveVariableScope(ctx *FlowContext, id string, scope func() (string, error)) (null.String, error) {
	scoped, err := ctx.Variable.VariableScoped(ctx, id)
	if err != nil || !scoped {
		return null.String{}, err
	}

	s, err := scope()
	if err != nil {
		return null.String{}, err
	}
	if s == "" {
		return null.String{}, errEmptyVariableScope
	}
	return null.StringFrom(s), nil
}

// variableScope evaluates the scope of a variable block.
func variableScope(n *CompiledFlowNode, ctx *FlowContext) (null.String, error) {
	return resolveVariableScope(ctx, n.Data.VariableID, func() (string, error) {
		scope, err := ctx.EvalTemplate(n.Data.VariableScope)
		return scope.String(), err
	})
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

// executeActionVariableList stores the values of a variable as a list of
// objects with the scope, the value and when it was last updated, like a
// leaderboard of the highest values.
func executeActionVariableList(n *CompiledFlowNode, ctx *FlowContext) error {
	limit, err := variableListNumber(ctx, n.Data.VariableLimit, defaultVariableListLimit, 1, maxVariableListLimit)
	if err != nil {
		return fmt.Errorf("the number of values must be a whole number from 1 to %d", maxVariableListLimit)
	}
	skip, err := variableListNumber(ctx, n.Data.VariableSkip, 0, 0, maxVariableListSkip)
	if err != nil {
		return fmt.Errorf("the number of values to skip must be a whole number from 0 to %d", maxVariableListSkip)
	}

	// Unscoped variables have one value, which a prefix can't match.
	var prefix string
	scoped, err := ctx.Variable.VariableScoped(ctx, n.Data.VariableID)
	if err != nil {
		return err
	}
	if scoped {
		value, err := ctx.EvalTemplate(n.Data.VariableScopePrefix)
		if err != nil {
			return err
		}
		prefix = value.String()
	}

	sort := n.Data.VariableSort
	if sort == "" {
		sort = provider.VariableListSortValueDesc
	}

	values, err := ctx.Variable.VariableValues(ctx, n.Data.VariableID, provider.VariableListOptions{
		ScopePrefix: prefix,
		Sort:        sort,
		Limit:       limit,
		Skip:        skip,
	})
	if err != nil {
		return err
	}

	ctx.StoreNodeResult(n, variableValueRows(values))
	return n.ExecuteChildren(ctx)
}

// variableValueRows are values of a variable as a list of objects with the
// scope, the value and when it was last updated.
func variableValueRows(values []provider.VariableValue) thing.Thing {
	items := make([]thing.Thing, len(values))
	for i, v := range values {
		scope := thing.Null
		if v.Scope.Valid {
			scope = thing.NewString(v.Scope.String)
		}
		items[i] = thing.NewObject(map[string]thing.Thing{
			"scope":      scope,
			"value":      v.Value,
			"updated_at": thing.NewInt(v.UpdatedAt.Unix()),
		})
	}
	return thing.NewArray(items)
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
