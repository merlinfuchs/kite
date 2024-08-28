package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type VariableStore interface {
	VariablesByApp(ctx context.Context, appID string) ([]*model.Variable, error)
	CountVariablesByApp(ctx context.Context, appID string) (int, error)
	Variable(ctx context.Context, id string) (*model.Variable, error)
	VariableScope(ctx context.Context, id string) (model.VariableScope, error)
	VariableByName(ctx context.Context, appID, name string) (*model.Variable, error)
	CreateVariable(ctx context.Context, variable *model.Variable) (*model.Variable, error)
	UpdateVariable(ctx context.Context, variable *model.Variable) (*model.Variable, error)
	DeleteVariable(ctx context.Context, id string) error
}

type VariableValueStore interface {
	VariableValues(ctx context.Context, variableID string) ([]*model.VariableValue, error)
	VariableValue(ctx context.Context, variableID string, scope model.VariableValueScope) (*model.VariableValue, error)
	SetVariableValue(ctx context.Context, value model.VariableValue, scope model.VariableValueScope) error
	DeleteVariableValue(ctx context.Context, variableID string, scope model.VariableValueScope) error
	DeleteAllVariableValues(ctx context.Context, variableID string) error
}
