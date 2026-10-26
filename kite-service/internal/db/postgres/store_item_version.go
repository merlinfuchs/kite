package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

// autoSaveMergeWindow is how long the auto-saves of a user go into the same
// version, so a long editing session doesn't push all older versions out.
const autoSaveMergeWindow = 10 * time.Minute

// itemVersionTarget is the command, event listener or message a version
// belongs to.
type itemVersionTarget struct {
	appID      string
	targetType model.ItemVersionTargetType
	id         string
}

// recordItemChange adds the saved content to the history of its target,
// unless the change left it as it was. The content from before the change is
// added first if the target has no versions yet, so the first change can be
// undone too.
func recordItemChange(
	ctx context.Context,
	q *pgmodel.Queries,
	target itemVersionTarget,
	previous any,
	previousAt time.Time,
	saved any,
	change store.ItemChange,
) error {
	rawPrevious, err := json.Marshal(previous)
	if err != nil {
		return err
	}
	rawSaved, err := json.Marshal(saved)
	if err != nil {
		return err
	}
	if bytes.Equal(rawPrevious, rawSaved) {
		return nil
	}

	now := time.Now().UTC()
	newest, err := q.GetNewestItemVersion(ctx, target.id)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		err := createItemVersion(ctx, q, target, model.ItemVersionSourceInitial, "", rawPrevious, nil, previousAt)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	case change.Source == model.ItemVersionSourceAutoSave &&
		newest.Source == string(model.ItemVersionSourceAutoSave) &&
		newest.CreatorUserID.String == change.UserID &&
		now.Sub(newest.CreatedAt.Time) < autoSaveMergeWindow:
		return q.UpdateItemVersion(ctx, pgmodel.UpdateItemVersionParams{
			ID:        newest.ID,
			Content:   rawSaved,
			CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
		})
	}

	if err := createItemVersion(ctx, q, target, change.Source, change.UserID, rawSaved, nil, now); err != nil {
		return err
	}
	return q.DeleteOldItemVersions(ctx, pgmodel.DeleteOldItemVersionsParams{
		TargetID:  target.id,
		KeepCount: model.MaxItemVersions,
	})
}

// recordItemDelete adds the last content of a deleted target to its history,
// with the rest of the target in targetData so it can be restored.
func recordItemDelete(
	ctx context.Context,
	q *pgmodel.Queries,
	target itemVersionTarget,
	rawContent []byte,
	targetData any,
	userID string,
) error {
	rawTargetData, err := json.Marshal(targetData)
	if err != nil {
		return err
	}

	err = createItemVersion(ctx, q, target, model.ItemVersionSourceDelete, userID, rawContent, rawTargetData, time.Now().UTC())
	if err != nil {
		return err
	}
	return q.DeleteOldItemVersions(ctx, pgmodel.DeleteOldItemVersionsParams{
		TargetID:  target.id,
		KeepCount: model.MaxItemVersions,
	})
}

func createItemVersion(
	ctx context.Context,
	q *pgmodel.Queries,
	target itemVersionTarget,
	source model.ItemVersionSource,
	userID string,
	rawContent []byte,
	rawTargetData []byte,
	createdAt time.Time,
) error {
	return q.CreateItemVersion(ctx, pgmodel.CreateItemVersionParams{
		ID:            util.UniqueID(),
		AppID:         target.appID,
		TargetType:    string(target.targetType),
		TargetID:      target.id,
		Source:        string(source),
		CreatorUserID: pgtype.Text{String: userID, Valid: userID != ""},
		Content:       rawContent,
		TargetData:    rawTargetData,
		CreatedAt:     pgtype.Timestamp{Time: createdAt.UTC(), Valid: true},
	})
}

func (c *Client) ItemVersion(ctx context.Context, appID string, id string) (*model.ItemVersion, error) {
	row, err := c.Q.GetItemVersion(ctx, pgmodel.GetItemVersionParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.ItemVersion{
		ID:            row.ID,
		AppID:         row.AppID,
		TargetType:    model.ItemVersionTargetType(row.TargetType),
		TargetID:      row.TargetID,
		Source:        model.ItemVersionSource(row.Source),
		CreatorUserID: null.NewString(row.CreatorUserID.String, row.CreatorUserID.Valid),
		Content:       row.Content,
		CreatedAt:     row.CreatedAt.Time,
	}, nil
}

func (c *Client) ItemVersionsByTarget(ctx context.Context, appID string, targetID string, limit int) ([]*model.ItemVersion, error) {
	rows, err := c.Q.GetItemVersionsByTarget(ctx, pgmodel.GetItemVersionsByTargetParams{
		AppID:    appID,
		TargetID: targetID,
		MaxCount: int32(limit),
	})
	if err != nil {
		return nil, err
	}

	versions := make([]*model.ItemVersion, len(rows))
	for i, row := range rows {
		versions[i] = &model.ItemVersion{
			ID:                 row.ID,
			AppID:              row.AppID,
			TargetType:         model.ItemVersionTargetType(row.TargetType),
			TargetID:           row.TargetID,
			Source:             model.ItemVersionSource(row.Source),
			CreatorUserID:      null.NewString(row.CreatorUserID.String, row.CreatorUserID.Valid),
			CreatorDisplayName: null.NewString(row.CreatorDisplayName.String, row.CreatorDisplayName.Valid),
			CreatedAt:          row.CreatedAt.Time,
		}
	}

	return versions, nil
}

func (c *Client) DeleteDeletedItemVersions(ctx context.Context, deletedBefore time.Time, batchSize int) (int64, error) {
	return c.Q.DeleteDeletedItemVersions(ctx, pgmodel.DeleteDeletedItemVersionsParams{
		DeletedBefore: pgtype.Timestamp{Time: deletedBefore.UTC(), Valid: true},
		BatchSize:     int32(batchSize),
	})
}
