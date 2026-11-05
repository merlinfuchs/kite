package provider

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
)

type VariableOperation string

const (
	VariableOperationOverwrite VariableOperation = "overwrite"
	VariableOperationAppend    VariableOperation = "append"
	VariableOperationPrepend   VariableOperation = "prepend"
	VariableOperationIncrement VariableOperation = "increment"
	VariableOperationDecrement VariableOperation = "decrement"
)

func (o VariableOperation) IsOverwrite() bool {
	return o == VariableOperationOverwrite || o == ""
}

// VariableProvider provides access to user-defined variables and their values.
type VariableProvider interface {
	UpdateVariable(ctx context.Context, id string, scope null.String, operation VariableOperation, value thing.Thing) (thing.Thing, error)
	Variable(ctx context.Context, id string, scope null.String) (thing.Thing, error)
	DeleteVariable(ctx context.Context, id string, scope null.String) error
	// VariableScoped reports whether the variable has a value per scope. It's
	// false for unknown variables.
	VariableScoped(ctx context.Context, id string) (bool, error)
	// VariableValues lists the values of a variable in the order of opts.Sort.
	VariableValues(ctx context.Context, id string, opts VariableListOptions) ([]VariableValue, error)
}

type VariableListSort string

const (
	// Values that are numbers, or text of a number, sort as numbers. Others
	// come last.
	VariableListSortValueDesc   VariableListSort = "value_desc"
	VariableListSortValueAsc    VariableListSort = "value_asc"
	VariableListSortUpdatedDesc VariableListSort = "updated_desc"
	VariableListSortUpdatedAsc  VariableListSort = "updated_asc"
)

type VariableListOptions struct {
	// Only values whose scope starts with it, all values if empty.
	ScopePrefix string
	Sort        VariableListSort
	Limit       int
}

type VariableValue struct {
	Scope     null.String
	Value     thing.Thing
	UpdatedAt time.Time
}

type MockVariableProvider struct{}

func (p *MockVariableProvider) UpdateVariable(ctx context.Context, id string, scope null.String, operation VariableOperation, value thing.Thing) (thing.Thing, error) {
	return thing.Null, nil
}

func (p *MockVariableProvider) Variable(ctx context.Context, id string, scope null.String) (thing.Thing, error) {
	return thing.Null, nil
}

func (p *MockVariableProvider) DeleteVariable(ctx context.Context, id string, scope null.String) error {
	return nil
}

func (p *MockVariableProvider) VariableScoped(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (p *MockVariableProvider) VariableValues(ctx context.Context, id string, opts VariableListOptions) ([]VariableValue, error) {
	return nil, nil
}
