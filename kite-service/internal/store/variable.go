package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type VariableStore interface {
	VariablesByApp(ctx context.Context, appID string) ([]*model.Variable, error)
	// VariablesByAppWithoutTotals is VariablesByApp without counting the
	// values of each variable, which is slow for variables with many.
	VariablesByAppWithoutTotals(ctx context.Context, appID string) ([]*model.Variable, error)
	CountVariablesByApp(ctx context.Context, appID string) (int, error)
	Variable(ctx context.Context, id string) (*model.Variable, error)
	VariableByName(ctx context.Context, appID, name string) (*model.Variable, error)
	// CreateVariable and UpdateVariable return ErrAlreadyExists if the app has
	// another variable with the name.
	CreateVariable(ctx context.Context, variable *model.Variable) (*model.Variable, error)
	UpdateVariable(ctx context.Context, variable *model.Variable) (*model.Variable, error)
	DeleteVariable(ctx context.Context, id string) error
}

type VariableValueStore interface {
	VariableValues(ctx context.Context, appID string, variableID string) ([]*model.VariableValue, error)
	// SearchVariableValues returns a page of the variable's values whose scope
	// contains search, most recently updated first, and how many match in total.
	SearchVariableValues(ctx context.Context, appID string, variableID string, search string, limit int, offset int) ([]*model.VariableValue, int, error)
	// ListVariableValues returns up to limit values of the variable whose scope
	// starts with scopePrefix, sorted by sort, one of value_desc, value_asc,
	// updated_desc and updated_asc, after skipping offset of them.
	ListVariableValues(ctx context.Context, appID string, variableID string, scopePrefix string, sort string, limit int, offset int) ([]*model.VariableValue, error)
	VariableValue(ctx context.Context, appID string, variableID string, scope null.String) (*model.VariableValue, error)
	SetVariableValue(ctx context.Context, appID string, value model.VariableValue) error
	UpdateVariableValue(ctx context.Context, appID string, operation model.VariableValueOperation, value model.VariableValue) (*model.VariableValue, error)
	DeleteVariableValue(ctx context.Context, appID string, variableID string, scope null.String) error
	DeleteAllVariableValues(ctx context.Context, appID string, variableID string) error
}
