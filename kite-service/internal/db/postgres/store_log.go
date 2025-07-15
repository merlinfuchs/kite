package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreateLogEntry(ctx context.Context, entry model.LogEntry) error {
	err := c.Q.CreateLogEntry(ctx, pgmodel.CreateLogEntryParams{
		AppID:   entry.AppID,
		Level:   string(entry.Level),
		Message: entry.Message,
		CommandID: pgtype.Text{
			String: entry.CommandID.String,
			Valid:  entry.CommandID.Valid,
		},
		EventListenerID: pgtype.Text{
			String: entry.EventListenerID.String,
			Valid:  entry.EventListenerID.Valid,
		},
		MessageID: pgtype.Text{
			String: entry.MessageID.String,
			Valid:  entry.MessageID.Valid,
		},
		CreatedAt: pgtype.Timestamp{
			Time:  entry.CreatedAt,
			Valid: true,
		},
	})
	return err
}

func (c *Client) LogEntriesByApp(ctx context.Context, appID string, beforeID int64, limit int) ([]*model.LogEntry, error) {
	rows, err := c.Q.GetLogEntriesByApp(ctx, pgmodel.GetLogEntriesByAppParams{
		AppID: appID,
		Limit: int32(limit),
		BeforeID: pgtype.Int8{
			Int64: beforeID,
			Valid: beforeID != 0,
		},
	})
	if err != nil {
		return nil, err
	}

	var res []*model.LogEntry
	for _, row := range rows {
		res = append(res, rowToLogEntry(row))
	}

	return res, nil
}

func (c *Client) LogEntriesByCommand(ctx context.Context, appID string, commandID string, beforeID int64, limit int) ([]*model.LogEntry, error) {
	rows, err := c.Q.GetLogEntriesByCommand(ctx, pgmodel.GetLogEntriesByCommandParams{
		AppID: appID,
		CommandID: pgtype.Text{
			String: commandID,
			Valid:  true,
		},
		Limit: int32(limit),
		BeforeID: pgtype.Int8{
			Int64: beforeID,
			Valid: beforeID != 0,
		},
	})
	if err != nil {
		return nil, err
	}

	var res []*model.LogEntry
	for _, row := range rows {
		res = append(res, rowToLogEntry(row))
	}

	return res, nil
}

func (c *Client) LogEntriesByEvent(ctx context.Context, appID string, eventID string, beforeID int64, limit int) ([]*model.LogEntry, error) {
	rows, err := c.Q.GetLogEntriesByEvent(ctx, pgmodel.GetLogEntriesByEventParams{
		AppID: appID,
		EventListenerID: pgtype.Text{
			String: eventID,
			Valid:  true,
		},
		Limit: int32(limit),
		BeforeID: pgtype.Int8{
			Int64: beforeID,
			Valid: beforeID != 0,
		},
	})
	if err != nil {
		return nil, err
	}

	var res []*model.LogEntry
	for _, row := range rows {
		res = append(res, rowToLogEntry(row))
	}

	return res, nil
}

func (c *Client) LogEntriesByMessage(ctx context.Context, appID string, messageID string, beforeID int64, limit int) ([]*model.LogEntry, error) {
	rows, err := c.Q.GetLogEntriesByMessage(ctx, pgmodel.GetLogEntriesByMessageParams{
		AppID: appID,
		MessageID: pgtype.Text{
			String: messageID,
			Valid:  true,
		},
		Limit: int32(limit),
		BeforeID: pgtype.Int8{
			Int64: beforeID,
			Valid: beforeID != 0,
		},
	})
	if err != nil {
		return nil, err
	}

	var res []*model.LogEntry
	for _, row := range rows {
		res = append(res, rowToLogEntry(row))
	}

	return res, nil
}

func (c *Client) LogSummary(ctx context.Context, appID string, start time.Time, end time.Time) (*model.LogSummary, error) {
	res, err := c.Q.GetLogSummary(ctx, pgmodel.GetLogSummaryParams{
		AppID:   appID,
		StartAt: pgtype.Timestamp{Time: start, Valid: true},
		EndAt:   pgtype.Timestamp{Time: end, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &model.LogSummary{
		TotalEntries:  res.TotalEntries,
		TotalErrors:   res.TotalErrors,
		TotalWarnings: res.TotalWarnings,
		TotalInfos:    res.TotalInfos,
		TotalDebugs:   res.TotalDebugs,
	}, nil
}

func rowToLogEntry(row pgmodel.Log) *model.LogEntry {
	return &model.LogEntry{
		ID:              row.ID,
		AppID:           row.AppID,
		Level:           model.LogLevel(row.Level),
		Message:         row.Message,
		CommandID:       null.NewString(row.CommandID.String, row.CommandID.Valid),
		EventListenerID: null.NewString(row.EventListenerID.String, row.EventListenerID.Valid),
		MessageID:       null.NewString(row.MessageID.String, row.MessageID.Valid),
		CreatedAt:       row.CreatedAt.Time,
	}
}
