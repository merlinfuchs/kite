package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
)

// variableArg names the stored variable a call uses, by the name the block
// gives it, and the scope of its value.
type variableArg struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

// codeVariableID is the ID of the stored variable the code calls name.
func codeVariableID(call operationCall, name string) (string, error) {
	if name == "" {
		return "", errors.New("missing parameter name")
	}
	id, ok := call.variables[name]
	if !ok {
		return "", fmt.Errorf("the block has no stored variable named %s, add it under Stored variables", name)
	}
	if id == "" {
		return "", fmt.Errorf("no variable is picked for %s under Stored variables of the block", name)
	}
	return id, nil
}

// startVariableCall resolves the variable and scope of a call like the
// blocks do, and charges for it.
func startVariableCall(call operationCall, v variableArg) (string, null.String, error) {
	id, err := codeVariableID(call, v.Name)
	if err != nil {
		return "", null.String{}, err
	}
	scope, err := resolveVariableScope(call.ctx, id, func() (string, error) { return v.Scope, nil })
	if errors.Is(err, errEmptyVariableScope) {
		return "", null.String{}, fmt.Errorf("%s has a value per scope, pass the scope, like a user ID", v.Name)
	} else if err != nil {
		return "", null.String{}, err
	}
	return id, scope, call.charge()
}

func variableGetOperation(call operationCall) (any, error) {
	var arg variableArg
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	id, scope, err := startVariableCall(call, arg)
	if err != nil {
		return nil, err
	}
	v, err := call.ctx.Variable.Variable(call.ctx, id, scope)
	if errors.Is(err, provider.ErrNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return eval.ThingJSON(v)
}

func variableSetOperation(call operationCall) (any, error) {
	var arg struct {
		variableArg
		Value json.RawMessage `json:"value"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	var value any
	if len(arg.Value) == 0 {
		return nil, errors.New("missing parameter value")
	} else if err := json.Unmarshal(arg.Value, &value); err != nil {
		return nil, errors.New("invalid value for parameter value")
	}

	id, scope, err := startVariableCall(call, arg.variableArg)
	if err != nil {
		return nil, err
	}
	return updateVariable(call, id, scope, provider.VariableOperationOverwrite, thing.NewFromJSONValue(value))
}

func variableIncrementOperation(call operationCall) (any, error) {
	var arg struct {
		variableArg
		By *float64 `json:"by"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	id, scope, err := startVariableCall(call, arg.variableArg)
	if err != nil {
		return nil, err
	}
	by := 1.0
	if arg.By != nil {
		by = *arg.By
	}
	return updateVariable(call, id, scope, provider.VariableOperationIncrement, thing.NewGuessTypeWithFallback(by))
}

func updateVariable(call operationCall, id string, scope null.String, op provider.VariableOperation, value thing.Thing) (any, error) {
	v, err := call.ctx.Variable.UpdateVariable(call.ctx, id, scope, op, value)
	if err != nil {
		return nil, err
	}
	return eval.ThingJSON(v)
}

func variableDeleteOperation(call operationCall) (any, error) {
	var arg variableArg
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	id, scope, err := startVariableCall(call, arg)
	if err != nil {
		return nil, err
	}
	return nil, call.ctx.Variable.DeleteVariable(call.ctx, id, scope)
}

// variableListOperation lists the values of a variable like the List stored
// variable values block.
func variableListOperation(call operationCall) (any, error) {
	var arg struct {
		Name        string                    `json:"name"`
		ScopePrefix string                    `json:"scope_prefix"`
		Sort        provider.VariableListSort `json:"sort"`
		Limit       *int                      `json:"limit"`
		Skip        int                       `json:"skip"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	if arg.Sort == "" {
		arg.Sort = provider.VariableListSortValueDesc
	}
	if !slices.Contains(provider.VariableListSorts, arg.Sort) {
		sorts := make([]string, len(provider.VariableListSorts))
		for i, s := range provider.VariableListSorts {
			sorts[i] = string(s)
		}
		return nil, fmt.Errorf("invalid value for parameter sort: must be one of %s", strings.Join(sorts, ", "))
	}
	limit := defaultVariableListLimit
	if arg.Limit != nil {
		limit = *arg.Limit
	}
	if limit < 1 || limit > maxVariableListLimit {
		return nil, fmt.Errorf("invalid value for parameter limit: must be from 1 to %d", maxVariableListLimit)
	}
	if arg.Skip < 0 || arg.Skip > maxVariableListSkip {
		return nil, fmt.Errorf("invalid value for parameter skip: must be from 0 to %d", maxVariableListSkip)
	}

	id, err := codeVariableID(call, arg.Name)
	if err != nil {
		return nil, err
	}
	// Unscoped variables have one value, which a prefix can't match.
	scoped, err := call.ctx.Variable.VariableScoped(call.ctx, id)
	if err != nil {
		return nil, err
	}
	if !scoped {
		arg.ScopePrefix = ""
	}
	if err := call.charge(); err != nil {
		return nil, err
	}

	values, err := call.ctx.Variable.VariableValues(call.ctx, id, provider.VariableListOptions{
		ScopePrefix: arg.ScopePrefix,
		Sort:        arg.Sort,
		Limit:       limit,
		Skip:        arg.Skip,
	})
	if err != nil {
		return nil, err
	}
	return eval.ThingJSON(variableValueRows(values))
}
