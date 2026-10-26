package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// FlowChange is who or what changes a flow, for its version history. Stores
// record a version for every change of a flow and for every delete.
type FlowChange struct {
	Source model.FlowVersionSource
	UserID string
}

type FlowVersionStore interface {
	FlowVersion(ctx context.Context, appID string, id string) (*model.FlowVersion, error)
	// FlowVersionsByTarget returns the newest versions first, without their flow.
	FlowVersionsByTarget(ctx context.Context, appID string, targetID string, limit int) ([]*model.FlowVersion, error)
	// DeleteDeletedFlowVersions deletes up to batchSize versions of commands
	// and event listeners deleted before deletedBefore.
	DeleteDeletedFlowVersions(ctx context.Context, deletedBefore time.Time, batchSize int) (int64, error)
}
