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
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// autoSaveMergeWindow is how long the auto-saves of a user go into the same
// version, so a long editing session doesn't push all older versions out.
const autoSaveMergeWindow = 10 * time.Minute

// flowVersionTarget is the command or event listener a flow belongs to.
type flowVersionTarget struct {
	appID      string
	targetType model.FlowVersionTargetType
	id         string
}

// recordFlowChange adds the saved flow to the history of its target, unless
// the change left it as it was. The flow from before the change is added
// first if the target has no versions yet, so the first change can be undone
// too.
func recordFlowChange(
	ctx context.Context,
	q *pgmodel.Queries,
	target flowVersionTarget,
	previous flow.FlowData,
	previousAt time.Time,
	saved flow.FlowData,
	change store.FlowChange,
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
	newest, err := q.GetNewestFlowVersion(ctx, target.id)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		err := createFlowVersion(ctx, q, target, model.FlowVersionSourceInitial, "", rawPrevious, nil, previousAt)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	case change.Source == model.FlowVersionSourceAutoSave &&
		newest.Source == string(model.FlowVersionSourceAutoSave) &&
		newest.CreatorUserID.String == change.UserID &&
		now.Sub(newest.CreatedAt.Time) < autoSaveMergeWindow:
		return q.UpdateFlowVersion(ctx, pgmodel.UpdateFlowVersionParams{
			ID:         newest.ID,
			FlowSource: rawSaved,
			CreatedAt:  pgtype.Timestamp{Time: now, Valid: true},
		})
	}

	if err := createFlowVersion(ctx, q, target, change.Source, change.UserID, rawSaved, nil, now); err != nil {
		return err
	}
	return q.DeleteOldFlowVersions(ctx, pgmodel.DeleteOldFlowVersionsParams{
		TargetID:  target.id,
		KeepCount: model.MaxFlowVersions,
	})
}

// recordFlowDelete adds the last flow of a deleted target to its history,
// with the rest of the target in targetData so it can be restored.
func recordFlowDelete(
	ctx context.Context,
	q *pgmodel.Queries,
	target flowVersionTarget,
	rawFlowSource []byte,
	targetData any,
	userID string,
) error {
	rawTargetData, err := json.Marshal(targetData)
	if err != nil {
		return err
	}

	err = createFlowVersion(ctx, q, target, model.FlowVersionSourceDelete, userID, rawFlowSource, rawTargetData, time.Now().UTC())
	if err != nil {
		return err
	}
	return q.DeleteOldFlowVersions(ctx, pgmodel.DeleteOldFlowVersionsParams{
		TargetID:  target.id,
		KeepCount: model.MaxFlowVersions,
	})
}

func createFlowVersion(
	ctx context.Context,
	q *pgmodel.Queries,
	target flowVersionTarget,
	source model.FlowVersionSource,
	userID string,
	rawFlowSource []byte,
	rawTargetData []byte,
	createdAt time.Time,
) error {
	return q.CreateFlowVersion(ctx, pgmodel.CreateFlowVersionParams{
		ID:            util.UniqueID(),
		AppID:         target.appID,
		TargetType:    string(target.targetType),
		TargetID:      target.id,
		Source:        string(source),
		CreatorUserID: pgtype.Text{String: userID, Valid: userID != ""},
		FlowSource:    rawFlowSource,
		TargetData:    rawTargetData,
		CreatedAt:     pgtype.Timestamp{Time: createdAt.UTC(), Valid: true},
	})
}

func (c *Client) FlowVersion(ctx context.Context, appID string, id string) (*model.FlowVersion, error) {
	row, err := c.Q.GetFlowVersion(ctx, pgmodel.GetFlowVersionParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	var flowSource flow.FlowData
	if err := json.Unmarshal(row.FlowSource, &flowSource); err != nil {
		return nil, err
	}

	return &model.FlowVersion{
		ID:            row.ID,
		AppID:         row.AppID,
		TargetType:    model.FlowVersionTargetType(row.TargetType),
		TargetID:      row.TargetID,
		Source:        model.FlowVersionSource(row.Source),
		CreatorUserID: null.NewString(row.CreatorUserID.String, row.CreatorUserID.Valid),
		FlowSource:    flowSource,
		CreatedAt:     row.CreatedAt.Time,
	}, nil
}

func (c *Client) FlowVersionsByTarget(ctx context.Context, appID string, targetID string, limit int) ([]*model.FlowVersion, error) {
	rows, err := c.Q.GetFlowVersionsByTarget(ctx, pgmodel.GetFlowVersionsByTargetParams{
		AppID:    appID,
		TargetID: targetID,
		MaxCount: int32(limit),
	})
	if err != nil {
		return nil, err
	}

	versions := make([]*model.FlowVersion, len(rows))
	for i, row := range rows {
		versions[i] = &model.FlowVersion{
			ID:                 row.ID,
			AppID:              row.AppID,
			TargetType:         model.FlowVersionTargetType(row.TargetType),
			TargetID:           row.TargetID,
			Source:             model.FlowVersionSource(row.Source),
			CreatorUserID:      null.NewString(row.CreatorUserID.String, row.CreatorUserID.Valid),
			CreatorDisplayName: null.NewString(row.CreatorDisplayName.String, row.CreatorDisplayName.Valid),
			CreatedAt:          row.CreatedAt.Time,
		}
	}

	return versions, nil
}

func (c *Client) DeleteDeletedFlowVersions(ctx context.Context, deletedBefore time.Time, batchSize int) (int64, error) {
	return c.Q.DeleteDeletedFlowVersions(ctx, pgmodel.DeleteDeletedFlowVersionsParams{
		DeletedBefore: pgtype.Timestamp{Time: deletedBefore.UTC(), Valid: true},
		BatchSize:     int32(batchSize),
	})
}
