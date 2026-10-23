package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type BundleInstallStore interface {
	BundleInstallsByApp(ctx context.Context, appID string) ([]*model.BundleInstall, error)
	CreateBundleInstall(ctx context.Context, install *model.BundleInstall) (*model.BundleInstall, error)
}
