package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

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
// undone too. Both contents are JSON as Postgres returns it, so equal content
// has equal bytes.
func recordItemChange(
	ctx context.Context,
	q *pgmodel.Queries,
	target itemVersionTarget,
	previous []byte,
	previousAt time.Time,
	saved []byte,
	change store.ItemChange,
) error {
	if bytes.Equal(previous, saved) {
		return nil
	}

	now := time.Now().UTC()
	newest, err := q.GetNewestItemVersion(ctx, target.id)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		initial := store.ItemChange{Source: model.ItemVersionSourceInitial}
		err := addItemVersion(ctx, q, target, initial, previous, nil, previousAt)
		if err != nil {
			return err
		}
	case err != nil:
		return fmt.Errorf("failed to get newest item version: %w", err)
	case mergesInto(newest, change, now):
		err := q.UpdateItemVersionContent(ctx, pgmodel.UpdateItemVersionContentParams{
			ID:        newest.ID,
			Content:   saved,
			UpdatedAt: pgtype.Timestamp{Time: now, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to update item version: %w", err)
		}
		return nil
	}

	return addItemVersion(ctx, q, target, change, saved, nil, now)
}

// mergesInto says whether a change updates the newest version instead of
// adding one: it has the same merge key, source and user, and came within the
// change's window of the version's first change.
func mergesInto(newest pgmodel.GetNewestItemVersionRow, change store.ItemChange, now time.Time) bool {
	return change.MergeKey != "" &&
		newest.MergeKey.String == change.MergeKey &&
		newest.Source == string(change.Source) &&
		newest.CreatorUserID.String == change.UserID &&
		now.Sub(newest.CreatedAt.Time) < change.MergeWithin
}

// recordItemDelete adds the last content of a deleted target to its history,
// with the rest of the target in targetData so it can be restored.
func recordItemDelete(
	ctx context.Context,
	q *pgmodel.Queries,
	target itemVersionTarget,
	content []byte,
	targetData any,
	userID string,
) error {
	rawTargetData, err := json.Marshal(targetData)
	if err != nil {
		return err
	}

	change := store.ItemChange{Source: model.ItemVersionSourceDelete, UserID: userID}
	return addItemVersion(ctx, q, target, change, content, rawTargetData, time.Now().UTC())
}

// addItemVersion adds a version and drops the oldest ones over the limit.
func addItemVersion(
	ctx context.Context,
	q *pgmodel.Queries,
	target itemVersionTarget,
	change store.ItemChange,
	content []byte,
	targetData []byte,
	createdAt time.Time,
) error {
	err := q.CreateItemVersion(ctx, pgmodel.CreateItemVersionParams{
		ID:            util.UniqueID(),
		AppID:         target.appID,
		TargetType:    string(target.targetType),
		TargetID:      target.id,
		Source:        string(change.Source),
		CreatorUserID: pgtype.Text{String: change.UserID, Valid: change.UserID != ""},
		Content:       content,
		TargetData:    targetData,
		MergeKey:      pgtype.Text{String: change.MergeKey, Valid: change.MergeKey != ""},
		CreatedAt:     pgtype.Timestamp{Time: createdAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create item version: %w", err)
	}

	err = q.DeleteOldItemVersions(ctx, pgmodel.DeleteOldItemVersionsParams{
		TargetID:  target.id,
		KeepCount: model.MaxItemVersions,
	})
	if err != nil {
		return fmt.Errorf("failed to delete old item versions: %w", err)
	}
	return nil
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
		UpdatedAt:     null.NewTime(row.UpdatedAt.Time, row.UpdatedAt.Valid),
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
			UpdatedAt:          null.NewTime(row.UpdatedAt.Time, row.UpdatedAt.Valid),
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
