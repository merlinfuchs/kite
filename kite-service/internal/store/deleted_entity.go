package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// DeletedEntityStore reads the tombstones the database records for deleted
// commands, event listeners and plugin instances.
type DeletedEntityStore interface {
	DeletedEntitiesSince(ctx context.Context, deletedSince time.Time) ([]*model.DeletedEntity, error)
	DeleteDeletedEntitiesBefore(ctx context.Context, deletedBefore time.Time, batchSize int) (int64, error)
}
