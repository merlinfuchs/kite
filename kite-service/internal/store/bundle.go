package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type BundleStore interface {
	// Bundle returns the bundle with the content of the version apps install,
	// without its listing.
	Bundle(ctx context.Context, id string) (*model.Bundle, error)
	// ListedBundle is Bundle for bundles that were published, with their
	// listing.
	ListedBundle(ctx context.Context, id string) (*model.Bundle, error)
	// OfficialBundles returns the listed official bundles in their order.
	OfficialBundles(ctx context.Context) ([]*model.Bundle, error)
	// UpsertOfficialBundle stores the bundle with its listing and makes its
	// content the version apps install. The content of a stored version is
	// replaced.
	UpsertOfficialBundle(ctx context.Context, bundle *model.Bundle) error
}
