package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type FunctionStore interface {
	FunctionsByApp(ctx context.Context, appID string) ([]*model.Function, error)
	CountFunctionsByApp(ctx context.Context, appID string) (int, error)
	// Function is scoped to the app, so a flow can't call another app's
	// functions.
	Function(ctx context.Context, appID string, id string) (*model.Function, error)
	CreateFunction(ctx context.Context, function *model.Function) (*model.Function, error)
	UpdateFunction(ctx context.Context, function *model.Function) (*model.Function, error)
	DeleteFunction(ctx context.Context, appID string, id string) error
}
