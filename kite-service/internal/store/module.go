package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ModuleStore interface {
	ModulesByApp(ctx context.Context, appID string) ([]*model.Module, error)
	CreateModule(ctx context.Context, module *model.Module) (*model.Module, error)
}
