package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// ItemChange is who or what changes an item, for its version history. Stores
// record a version for every change of a command, event listener or message
// and for every delete.
type ItemChange struct {
	Source model.ItemVersionSource
	UserID string
}

type ItemVersionStore interface {
	ItemVersion(ctx context.Context, appID string, id string) (*model.ItemVersion, error)
	// ItemVersionsByTarget returns the newest versions first, without their content.
	ItemVersionsByTarget(ctx context.Context, appID string, targetID string, limit int) ([]*model.ItemVersion, error)
	// DeleteDeletedItemVersions deletes up to batchSize versions of items
	// deleted before deletedBefore.
	DeleteDeletedItemVersions(ctx context.Context, deletedBefore time.Time, batchSize int) (int64, error)
}
