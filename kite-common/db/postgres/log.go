package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-common/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-common/model"
)

func (c *Client) CreateLogEntry(ctx context.Context, entry model.LogEntry) error {
	err := c.Q.CreateLogEntry(ctx, pgmodel.CreateLogEntryParams{
		AppID:   entry.AppID,
		Level:   string(entry.Level),
		Message: entry.Message,
		CreatedAt: pgtype.Timestamp{
			Time:  entry.CreatedAt,
			Valid: true,
		},
	})
	return err
}

func (c *Client) LogEntriesByApp(ctx context.Context, appID string, beforeID int64, limit int) ([]*model.LogEntry, error) {
	var rows []pgmodel.Log
	var err error

	if beforeID == 0 {
		rows, err = c.Q.GetLogEntriesByApp(ctx, pgmodel.GetLogEntriesByAppParams{
			AppID: appID,
			Limit: int32(limit),
		})
	} else {
		rows, err = c.Q.GetLogEntriesByAppBefore(ctx, pgmodel.GetLogEntriesByAppBeforeParams{
			AppID: appID,
			ID:    beforeID,
			Limit: int32(limit),
		})

	}

	if err != nil {
		return nil, err
	}

	var res []*model.LogEntry
	for _, row := range rows {
		res = append(res, rowToLogEntry(row))
	}

	return res, nil

}

func rowToLogEntry(row pgmodel.Log) *model.LogEntry {
	return &model.LogEntry{
		ID:        row.ID,
		AppID:     row.AppID,
		Level:     model.LogLevel(row.Level),
		Message:   row.Message,
		CreatedAt: row.CreatedAt.Time,
	}
}
