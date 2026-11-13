package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) UpdateEditorSessionTime(ctx context.Context, session model.EditorSession, activeSeconds int, now time.Time) error {
	return c.Q.UpdateEditorSessionTime(ctx, pgmodel.UpdateEditorSessionTimeParams{
		ID:              session.ID,
		AppID:           session.AppID,
		UserID:          session.UserID,
		CommandID:       pgtype.Text{String: session.CommandID.String, Valid: session.CommandID.Valid},
		EventListenerID: pgtype.Text{String: session.EventListenerID.String, Valid: session.EventListenerID.Valid},
		ActiveSeconds:   int32(min(activeSeconds, int(store.EditorSessionMaxTime.Seconds()))),
		SlackSeconds:    int32(store.EditorSessionTimeSlack.Seconds()),
		MaxSeconds:      int32(store.EditorSessionMaxTime.Seconds()),
		Now:             pgtype.Timestamp{Time: now.UTC(), Valid: true},
	})
}

func (c *Client) RecordEditorSessionSave(ctx context.Context, session model.EditorSession, now time.Time) error {
	return c.Q.RecordEditorSessionSave(ctx, pgmodel.RecordEditorSessionSaveParams{
		ID:              session.ID,
		AppID:           session.AppID,
		UserID:          session.UserID,
		CommandID:       pgtype.Text{String: session.CommandID.String, Valid: session.CommandID.Valid},
		EventListenerID: pgtype.Text{String: session.EventListenerID.String, Valid: session.EventListenerID.Valid},
		Now:             pgtype.Timestamp{Time: now.UTC(), Valid: true},
	})
}

func (c *Client) EditorUsageByApp(ctx context.Context, appID string, since time.Time) ([]*model.EditorUsage, error) {
	rows, err := c.Q.GetEditorUsageByApp(ctx, pgmodel.GetEditorUsageByAppParams{
		AppID: appID,
		Since: pgtype.Timestamp{Time: since.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	res := make([]*model.EditorUsage, len(rows))
	for i, row := range rows {
		res[i] = &model.EditorUsage{
			CommandID:       null.NewString(row.CommandID.String, row.CommandID.Valid),
			EventListenerID: null.NewString(row.EventListenerID.String, row.EventListenerID.Valid),
			ActiveSeconds:   int(row.ActiveSeconds),
			SaveCount:       int(row.SaveCount),
			LastSavedAt:     null.NewTime(row.LastSavedAt.Time, row.LastSavedAt.Valid),
		}
	}
	return res, nil
}

func (c *Client) DeleteOldEditorSessions(ctx context.Context, createdBefore time.Time, batchSize int) (int64, error) {
	return c.Q.DeleteOldEditorSessions(ctx, pgmodel.DeleteOldEditorSessionsParams{
		CreatedBefore: pgtype.Timestamp{Time: createdBefore.UTC(), Valid: true},
		BatchSize:     int32(batchSize),
	})
}
