package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type BundleInstallStore interface {
	BundleInstallsByApp(ctx context.Context, appID string) ([]*model.BundleInstall, error)
	BundleInstall(ctx context.Context, appID string, id string) (*model.BundleInstall, error)
	// InstalledContentsByApp returns the commands, event listeners, messages
	// and variables of the app that installs created.
	InstalledContentsByApp(ctx context.Context, appID string) (BundleInstallContents, error)
	// CreateBundleInstall creates the install together with its contents, or
	// nothing if one of them fails. It returns ErrAlreadyExists if the app
	// has the bundle already.
	CreateBundleInstall(ctx context.Context, install *model.BundleInstall, contents BundleInstallContents) (*model.BundleInstall, error)
	// UpdateBundleInstall saves the version, parameters and items of the
	// install together with the changes to its contents. It returns
	// ErrOutdated if the install was changed since fromUpdatedAt.
	UpdateBundleInstall(ctx context.Context, install *model.BundleInstall, fromUpdatedAt time.Time, changes BundleInstallChanges, userID string) (*model.BundleInstall, error)
	// ImportBundleContents creates the contents of an imported bundle, which
	// aren't linked to it, or nothing if one of them fails.
	ImportBundleContents(ctx context.Context, contents BundleInstallContents) error
	// DeleteBundleInstall deletes the install and the commands, event
	// listeners, messages and variables it created, except the messages and
	// variables in keepIDs, which stay without the install.
	DeleteBundleInstall(ctx context.Context, id string, keepIDs []string, userID string) error
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
