package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type BundleInstallStore interface {
	BundleInstallsByApp(ctx context.Context, appID string) ([]*model.BundleInstall, error)
	BundleInstall(ctx context.Context, appID string, id string) (*model.BundleInstall, error)
	// CreateBundleInstall creates the install together with its contents, or
	// nothing if one of them fails.
	CreateBundleInstall(ctx context.Context, install *model.BundleInstall, contents BundleInstallContents) (*model.BundleInstall, error)
	// UpdateBundleInstall saves the version, settings and items of the install
	// together with the changed contents, which must exist already.
	UpdateBundleInstall(ctx context.Context, install *model.BundleInstall, contents BundleInstallContents) (*model.BundleInstall, error)
	// DeleteBundleInstall deletes the install and the commands, event listeners
	// and messages it created.
	DeleteBundleInstall(ctx context.Context, id string) error
}

type BundleInstallContents struct {
	Commands       []*model.Command
	EventListeners []*model.EventListener
	Messages       []*model.Message
}
