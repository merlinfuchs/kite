package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type BundleStore interface {
	// BundleCurrentVersion returns the bundle with the content of the version
	// apps install, without its listing.
	BundleCurrentVersion(ctx context.Context, id string) (*model.Bundle, *model.BundleVersion, error)
	// ListedBundleCurrentVersion is BundleCurrentVersion for bundles that
	// were published, with their listing.
	ListedBundleCurrentVersion(ctx context.Context, id string) (*model.Bundle, *model.BundleVersion, error)
	// OfficialBundlesCurrentVersion returns the official bundles with their
	// listing.
	OfficialBundlesCurrentVersion(ctx context.Context) ([]*model.Bundle, []*model.BundleVersion, error)
	// UpsertOfficialBundle stores the bundle with its listing and makes the
	// version the one apps install. The content of a stored version is
	// replaced.
	UpsertOfficialBundle(ctx context.Context, bundle *model.Bundle, version *model.BundleVersion) error
}
