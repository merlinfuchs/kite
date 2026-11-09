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

// maxVariableListLimit is the most values a list of a variable has. A
// leaderboard or a list in a message needs far fewer.
const maxVariableListLimit = 100

const defaultVariableListLimit = 10

// maxVariableListSkip bounds how far into the values a list can page. The
// database sorts all of them either way, a far offset only adds to that.
const maxVariableListSkip = 10_000

// variableCall is a call of an operation of a stored variable. The variable
// blocks are presets of these operations, and code calls them by the names
// the block gives its variables, see the variables.* handlers below.
type variableCall struct {
	ctx *FlowContext
	id  string
	// scope returns the scope of the value. It's only called for scoped
	// variables, so blocks only evaluate theirs when it's needed.
	scope func() (string, error)
	// charge is operationCall.charge for code. Blocks are charged when they
	// start, so theirs does nothing.
	charge func() error
}

var errEmptyVariableScope = errors.New("the variable is scoped, but the scope is empty")

// start resolves the scope of the call and charges for it. The variable
// decides whether there is a scope: blocks and code keep theirs when it's
// switched to unscoped, and a scoped variable has no value without a scope.
func (c variableCall) start() (null.String, error) {
	scoped, err := c.ctx.Variable.VariableScoped(c.ctx, c.id)
	if err != nil {
		return null.String{}, err
	}
	var scope null.String
	if scoped {
		s, err := c.scope()
		if err != nil {
			return null.String{}, err
		}
		if s == "" {
			return null.String{}, errEmptyVariableScope
		}
		scope = null.StringFrom(s)
	}
	return scope, c.charge()
}

// getVariable is the value of the variable, null if it has none.
func getVariable(c variableCall) (thing.Thing, error) {
	scope, err := c.start()
	if err != nil {
		return thing.Null, err
	}
	v, err := c.ctx.Variable.Variable(c.ctx, c.id, scope)
	if errors.Is(err, provider.ErrNotFound) {
		return thing.Null, nil
	}
	return v, err
}

// updateVariable changes the value of the variable with op, and returns the
// new value.
func updateVariable(c variableCall, op provider.VariableOperation, value thing.Thing) (thing.Thing, error) {
	scope, err := c.start()
	if err != nil {
		return thing.Null, err
	}
	return c.ctx.Variable.UpdateVariable(c.ctx, c.id, scope, op, value)
}

func deleteVariable(c variableCall) error {
	scope, err := c.start()
	if err != nil {
		return err
	}
	return c.ctx.Variable.DeleteVariable(c.ctx, c.id, scope)
}

type variableListOptions struct {
	// Sort defaults to the highest values first.
	Sort  provider.VariableListSort
	Limit int
	Skip  int
	// scopePrefix returns the start of the scopes of the values to list. It's
	// only called for scoped variables, unscoped ones have one value, which a
	// prefix can't match.
	scopePrefix func() (string, error)
}

// listVariable lists the values of the variable as objects with the scope,
// the value and when it was last updated, like a leaderboard of the highest
// values.
func listVariable(c variableCall, opts variableListOptions) (thing.Thing, error) {
	if opts.Sort == "" {
		opts.Sort = provider.VariableListSortValueDesc
	}
	if !slices.Contains(provider.VariableListSorts, opts.Sort) {
		sorts := make([]string, len(provider.VariableListSorts))
		for i, s := range provider.VariableListSorts {
			sorts[i] = string(s)
		}
		return thing.Null, fmt.Errorf("invalid value for parameter sort: must be one of %s", strings.Join(sorts, ", "))
	}
	if opts.Limit < 1 || opts.Limit > maxVariableListLimit {
		return thing.Null, fmt.Errorf("invalid value for parameter limit: must be from 1 to %d", maxVariableListLimit)
	}
	if opts.Skip < 0 || opts.Skip > maxVariableListSkip {
		return thing.Null, fmt.Errorf("invalid value for parameter skip: must be from 0 to %d", maxVariableListSkip)
	}

	scoped, err := c.ctx.Variable.VariableScoped(c.ctx, c.id)
	if err != nil {
		return thing.Null, err
	}
	var prefix string
	if scoped {
		if prefix, err = opts.scopePrefix(); err != nil {
			return thing.Null, err
		}
	}
	if err := c.charge(); err != nil {
		return thing.Null, err
	}

	values, err := c.ctx.Variable.VariableValues(c.ctx, c.id, provider.VariableListOptions{
		ScopePrefix: prefix,
		Sort:        opts.Sort,
		Limit:       opts.Limit,
		Skip:        opts.Skip,
	})
	if err != nil {
		return thing.Null, err
	}

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
	return thing.NewArray(items), nil
}

// The handlers of the variables.* operations code calls.

// variableArg names the stored variable a call uses, by the name the block
// gives it, and the scope of its value.
type variableArg struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

// codeVariableCall is the variableCall of a call from code.
func codeVariableCall(call operationCall, v variableArg) (variableCall, error) {
	if v.Name == "" {
		return variableCall{}, errors.New("missing parameter name")
	}
	id, ok := call.variables[v.Name]
	if !ok {
		return variableCall{}, fmt.Errorf("the block has no stored variable named %s, add it under Stored variables", v.Name)
	}
	if id == "" {
		return variableCall{}, fmt.Errorf("no variable is picked for %s under Stored variables of the block", v.Name)
	}
	return variableCall{
		ctx:    call.ctx,
		id:     id,
		scope:  func() (string, error) { return v.Scope, nil },
		charge: call.charge,
	}, nil
}

// codeVariableResult is the result of a variables.* call as JSON for the
// code, with an error that says what the code can do about it.
func codeVariableResult(name string, v thing.Thing, err error) (any, error) {
	if errors.Is(err, errEmptyVariableScope) {
		return nil, fmt.Errorf("%s has a value per scope, pass the scope, like a user ID", name)
	} else if err != nil {
		return nil, err
	}
	return eval.ThingJSON(v)
}

func variableGetOperation(call operationCall) (any, error) {
	var arg variableArg
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	c, err := codeVariableCall(call, arg)
	if err != nil {
		return nil, err
	}
	v, err := getVariable(c)
	return codeVariableResult(arg.Name, v, err)
}

// variableValueOperation is a handler that updates the value of a variable
// with op and the value code passes, like variables.set.
func variableValueOperation(op provider.VariableOperation) operationHandler {
	return func(call operationCall) (any, error) {
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
		c, err := codeVariableCall(call, arg.variableArg)
		if err != nil {
			return nil, err
		}
		v, err := updateVariable(c, op, thing.NewFromJSONValue(value))
		return codeVariableResult(arg.Name, v, err)
	}
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
	c, err := codeVariableCall(call, arg.variableArg)
	if err != nil {
		return nil, err
	}
	v, err := updateVariable(c, provider.VariableOperationIncrement, thing.NewGuessTypeWithFallback(by))
	return codeVariableResult(arg.Name, v, err)
}

func variableDeleteOperation(call operationCall) (any, error) {
	var arg variableArg
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	c, err := codeVariableCall(call, arg)
	if err != nil {
		return nil, err
	}
	_, err = codeVariableResult(arg.Name, thing.Null, deleteVariable(c))
	return nil, err
}

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
	c, err := codeVariableCall(call, variableArg{Name: arg.Name})
	if err != nil {
		return nil, err
	}
	limit := defaultVariableListLimit
	if arg.Limit != nil {
		limit = *arg.Limit
	}
	v, err := listVariable(c, variableListOptions{
		Sort:        arg.Sort,
		Limit:       limit,
		Skip:        arg.Skip,
		scopePrefix: func() (string, error) { return arg.ScopePrefix, nil },
	})
	return codeVariableResult(arg.Name, v, err)
}
