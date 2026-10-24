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
	// together with the changes to its contents.
	UpdateBundleInstall(ctx context.Context, install *model.BundleInstall, changes BundleInstallChanges) (*model.BundleInstall, error)
	// ImportBundleContents creates the contents of an imported bundle, which
	// aren't linked to it, or nothing if one of them fails.
	ImportBundleContents(ctx context.Context, contents BundleInstallContents) error
	// DeleteBundleInstall deletes the install and the commands, event
	// listeners, messages and variables it created.
	DeleteBundleInstall(ctx context.Context, id string) error
}

type BundleInstallContents struct {
	Commands       []*model.Command
	EventListeners []*model.EventListener
	Messages       []*model.Message
	Variables      []*model.Variable
}

type BundleInstallChanges struct {
	Created BundleInstallContents
	// Updated must exist already.
	Updated BundleInstallContents
}
