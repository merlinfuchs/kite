package provider

import (
	"context"

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
	UpdateVariable(ctx context.Context, id string, scope null.String, operation VariableOperation, value thing.Any) (thing.Any, error)
	Variable(ctx context.Context, id string, scope null.String) (thing.Any, error)
	DeleteVariable(ctx context.Context, id string, scope null.String) error
}

type MockVariableProvider struct{}

func (p *MockVariableProvider) UpdateVariable(ctx context.Context, id string, scope null.String, operation VariableOperation, value thing.Any) (thing.Any, error) {
	return thing.Null, nil
}

func (p *MockVariableProvider) Variable(ctx context.Context, id string, scope null.String) (thing.Any, error) {
	return thing.Null, nil
}

func (p *MockVariableProvider) DeleteVariable(ctx context.Context, id string, scope null.String) error {
	return nil
}
