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
	// MergeKey, if set, makes changes with the same key update the newest
	// version instead of adding one, until MergeWithin after its first change.
	MergeKey    string
	MergeWithin time.Duration
}

const (
	// EditorSessionMergeWindow is how long the saves of one editor session
	// merge. Its undo covers the states in between while it's open, and a new
	// session starts a new version.
	EditorSessionMergeWindow = 30 * time.Minute
	// APITokenMergeWindow is how long the changes made with one API token
	// merge. An agent saves after every step, which would push all older
	// versions out, and undoing its session should undo all of it.
	APITokenMergeWindow = 10 * time.Minute
)

// EditorChange is a change a user saved in an editor. editorSession is made
// each time an editor is opened, so the saves while it's open merge. Saves
// without one, like restores, don't.
func EditorChange(userID string, editorSession string) ItemChange {
	return ItemChange{
		Source:      model.ItemVersionSourceEditor,
		UserID:      userID,
		MergeKey:    editorSession,
		MergeWithin: EditorSessionMergeWindow,
	}
}

// APITokenChange is a change a user made with an API token.
func APITokenChange(userID string, tokenID string) ItemChange {
	return ItemChange{
		Source:      model.ItemVersionSourceAPIToken,
		UserID:      userID,
		MergeKey:    tokenID,
		MergeWithin: APITokenMergeWindow,
	}
}

type ItemVersionStore interface {
	ItemVersion(ctx context.Context, appID string, id string) (*model.ItemVersion, error)
	// ItemVersionsByTarget returns the newest versions first, without their content.
	ItemVersionsByTarget(ctx context.Context, appID string, targetID string, limit int) ([]*model.ItemVersion, error)
	// DeleteDeletedItemVersions deletes the versions of up to batchSize items
	// deleted before deletedBefore, and returns how many versions it deleted.
	DeleteDeletedItemVersions(ctx context.Context, deletedBefore time.Time, batchSize int) (int64, error)
}
