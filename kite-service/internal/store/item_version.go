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

// EditorChange is a change a user made in the editor.
func EditorChange(userID string) ItemChange {
	return ItemChange{Source: model.ItemVersionSourceEditor, UserID: userID}
}

type ItemVersionStore interface {
	ItemVersion(ctx context.Context, appID string, id string) (*model.ItemVersion, error)
	// ItemVersionsByTarget returns the newest versions first, without their content.
	ItemVersionsByTarget(ctx context.Context, appID string, targetID string, limit int) ([]*model.ItemVersion, error)
	// DeleteDeletedItemVersions deletes the versions of up to batchSize items
	// deleted before deletedBefore, and returns how many versions it deleted.
	DeleteDeletedItemVersions(ctx context.Context, deletedBefore time.Time, batchSize int) (int64, error)
}
