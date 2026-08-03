package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type VariableStore interface {
	VariablesByApp(ctx context.Context, appID string) ([]*model.Variable, error)
	CountVariablesByApp(ctx context.Context, appID string) (int, error)
	Variable(ctx context.Context, id string) (*model.Variable, error)
	VariableByName(ctx context.Context, appID, name string) (*model.Variable, error)
	CreateVariable(ctx context.Context, variable *model.Variable) (*model.Variable, error)
	UpdateVariable(ctx context.Context, variable *model.Variable) (*model.Variable, error)
	DeleteVariable(ctx context.Context, id string) error
}

// VariableValueStore scopes every operation by app.
//
// variable_values rows carry no app_id, so the variable_id alone is the whole
// key; the app is what makes it a tenant-safe lookup. Callers reach this from
// flow node config, where the variable_id is attacker-authored, so the appID
// must come from the execution context rather than from the same flow.
type VariableValueStore interface {
	VariableValues(ctx context.Context, appID string, variableID string) ([]*model.VariableValue, error)
	VariableValue(ctx context.Context, appID string, variableID string, scope null.String) (*model.VariableValue, error)
	SetVariableValue(ctx context.Context, appID string, value model.VariableValue) error
	UpdateVariableValue(ctx context.Context, appID string, operation model.VariableValueOperation, value model.VariableValue) (*model.VariableValue, error)
	DeleteVariableValue(ctx context.Context, appID string, variableID string, scope null.String) error
	DeleteAllVariableValues(ctx context.Context, appID string, variableID string) error
}
