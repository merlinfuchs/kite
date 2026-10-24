package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type BundleInstallStore interface {
	BundleInstallsByApp(ctx context.Context, appID string) ([]*model.BundleInstall, error)
	// CreateBundleInstall creates the install together with its contents, or
	// nothing if one of them fails.
	CreateBundleInstall(ctx context.Context, install *model.BundleInstall, contents BundleInstallContents) (*model.BundleInstall, error)
}

type BundleInstallContents struct {
	Commands       []*model.Command
	EventListeners []*model.EventListener
}
