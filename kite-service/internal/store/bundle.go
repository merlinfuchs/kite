package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type BundleStore interface {
	// BundleCurrentVersion returns the bundle with the content of the version
	// apps install.
	BundleCurrentVersion(ctx context.Context, id string) (*model.Bundle, *model.BundleVersion, error)
	// UpsertOfficialBundle stores the bundle and makes the version the one
	// apps install. The content of a stored version is replaced.
	UpsertOfficialBundle(ctx context.Context, bundle *model.Bundle, version *model.BundleVersion) error
}
