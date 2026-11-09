package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
)

// codeVariableID is the ID of the stored variable a call names, as the block
// lets the code use it.
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

// codeVariable is codeVariableID with the scope of the call. Like the
// blocks, a scoped variable needs a scope and an unscoped one ignores it.
func codeVariable(call operationCall, name string, scope *string) (string, null.String, error) {
	id, err := codeVariableID(call, name)
	if err != nil {
		return "", null.String{}, err
	}
	scoped, err := call.ctx.Variable.VariableScoped(call.ctx, id)
	if err != nil || !scoped {
		return id, null.String{}, err
	}
	if scope == nil || *scope == "" {
		return "", null.String{}, fmt.Errorf("%s has a value per scope, pass the scope, like a user ID", name)
	}
	return id, null.StringFrom(*scope), nil
}

type variableArg struct {
	Name  string  `json:"name"`
	Scope *string `json:"scope"`
}

func variableGetOperation(call operationCall) (any, error) {
	var arg variableArg
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	id, scope, err := codeVariable(call, arg.Name, arg.Scope)
	if err != nil {
		return nil, err
	}
	if err := call.charge(); err != nil {
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
	if len(arg.Value) == 0 {
		return nil, errors.New("missing parameter value")
	}
	var value any
	if err := json.Unmarshal(arg.Value, &value); err != nil {
		return nil, fmt.Errorf("invalid value for parameter value")
	}
	return updateCodeVariable(call, arg.variableArg, provider.VariableOperationOverwrite, thing.NewFromJSONValue(value))
}

func variableIncrementOperation(call operationCall) (any, error) {
	var arg struct {
		variableArg
		By *float64 `json:"by"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	by := 1.0
	if arg.By != nil {
		by = *arg.By
	}
	return updateCodeVariable(call, arg.variableArg, provider.VariableOperationIncrement, thing.NewGuessTypeWithFallback(by))
}

func updateCodeVariable(call operationCall, arg variableArg, op provider.VariableOperation, value thing.Thing) (any, error) {
	id, scope, err := codeVariable(call, arg.Name, arg.Scope)
	if err != nil {
		return nil, err
	}
	if err := call.charge(); err != nil {
		return nil, err
	}

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
	id, scope, err := codeVariable(call, arg.Name, arg.Scope)
	if err != nil {
		return nil, err
	}
	if err := call.charge(); err != nil {
		return nil, err
	}
	return nil, call.ctx.Variable.DeleteVariable(call.ctx, id, scope)
}

var variableListSorts = []provider.VariableListSort{
	provider.VariableListSortValueDesc,
	provider.VariableListSortValueAsc,
	provider.VariableListSortUpdatedDesc,
	provider.VariableListSortUpdatedAsc,
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
	if !slices.Contains(variableListSorts, arg.Sort) {
		return nil, errors.New("invalid value for parameter sort: must be value_desc, value_asc, updated_desc or updated_asc")
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
	prefix := arg.ScopePrefix
	scoped, err := call.ctx.Variable.VariableScoped(call.ctx, id)
	if err != nil {
		return nil, err
	}
	if !scoped {
		prefix = ""
	}
	if err := call.charge(); err != nil {
		return nil, err
	}

	values, err := call.ctx.Variable.VariableValues(call.ctx, id, provider.VariableListOptions{
		ScopePrefix: prefix,
		Sort:        arg.Sort,
		Limit:       limit,
		Skip:        arg.Skip,
	})
	if err != nil {
		return nil, err
	}

	res := make([]map[string]any, len(values))
	for i, v := range values {
		value, err := eval.ThingJSON(v.Value)
		if err != nil {
			return nil, err
		}
		res[i] = map[string]any{
			"scope":      v.Scope.Ptr(),
			"value":      value,
			"updated_at": v.UpdatedAt.Unix(),
		}
	}
	return res, nil
}
