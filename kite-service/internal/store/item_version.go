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
	// MergeWithin makes changes of the same source and user within this time
	// of the newest version's first one update it instead of adding one.
	MergeWithin time.Duration
}

// EditorChange is a change a user made in the editor.
func EditorChange(userID string) ItemChange {
	return ItemChange{Source: model.ItemVersionSourceEditor, UserID: userID}
}

// APITokenMergeWindow is how long a user's changes with API tokens merge into
// one version. An agent saves after every step, which would push all older
// versions out, and undoing its session should undo all of it.
const APITokenMergeWindow = 10 * time.Minute

// APITokenChange is a change a user made with an API token.
func APITokenChange(userID string) ItemChange {
	return ItemChange{Source: model.ItemVersionSourceAPIToken, UserID: userID, MergeWithin: APITokenMergeWindow}
}

type ItemVersionStore interface {
	ItemVersion(ctx context.Context, appID string, id string) (*model.ItemVersion, error)
	// ItemVersionsByTarget returns the newest versions first, without their content.
	ItemVersionsByTarget(ctx context.Context, appID string, targetID string, limit int) ([]*model.ItemVersion, error)
	// DeleteDeletedItemVersions deletes the versions of up to batchSize items
	// deleted before deletedBefore, and returns how many versions it deleted.
	DeleteDeletedItemVersions(ctx context.Context, deletedBefore time.Time, batchSize int) (int64, error)
}
