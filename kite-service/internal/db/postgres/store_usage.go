package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreateUsageRecord(ctx context.Context, record model.UsageRecord) error {
	return c.Q.CreateUsageRecord(ctx, pgmodel.CreateUsageRecordParams{
		Type:            string(record.Type),
		AppID:           record.AppID,
		CommandID:       pgtype.Text{String: record.CommandID.String, Valid: record.CommandID.Valid},
		EventListenerID: pgtype.Text{String: record.EventListenerID.String, Valid: record.EventListenerID.Valid},
		MessageID:       pgtype.Text{String: record.MessageID.String, Valid: record.MessageID.Valid},
		CreditsUsed:     int32(record.CreditsUsed),
		CreatedAt:       pgtype.Timestamp{Time: record.CreatedAt, Valid: true},
	})
}

func (c *Client) UsageRecordsBetween(ctx context.Context, appID string, start time.Time, end time.Time) ([]model.UsageRecord, error) {
	rows, err := c.Q.GetUsageRecordsByAppBetween(ctx, pgmodel.GetUsageRecordsByAppBetweenParams{
		AppID:   appID,
		StartAt: pgtype.Timestamp{Time: start, Valid: true},
		EndAt:   pgtype.Timestamp{Time: end, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	records := make([]model.UsageRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, rowToUsageRecord(row))
	}

	return records, nil
}

func (c *Client) UsageCreditsUsedBetween(ctx context.Context, appID string, start time.Time, end time.Time) (int, error) {
	res, err := c.Q.GetUsageCreditsUsedByAppBetween(ctx, pgmodel.GetUsageCreditsUsedByAppBetweenParams{
		AppID:   appID,
		StartAt: pgtype.Timestamp{Time: start, Valid: true},
		EndAt:   pgtype.Timestamp{Time: end, Valid: true},
	})
	if err != nil {
		return 0, err
	}

	return int(res), nil
}

func (c *Client) UsageCreditsUsedByTypeBetween(ctx context.Context, appID string, start time.Time, end time.Time) ([]model.UsageCreditsUsedByType, error) {
	rows, err := c.Q.GetUsageCreditsUsedByTypeBetween(ctx, pgmodel.GetUsageCreditsUsedByTypeBetweenParams{
		AppID:   appID,
		StartAt: pgtype.Timestamp{Time: start, Valid: true},
		EndAt:   pgtype.Timestamp{Time: end, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	records := make([]model.UsageCreditsUsedByType, 0, len(rows))
	for _, row := range rows {
		records = append(records, model.UsageCreditsUsedByType{
			Type:        model.UsageRecordType(row.Type),
			CreditsUsed: int(row.Sum),
		})
	}

	return records, nil
}

func (c *Client) UsageCreditsUsedByDayBetween(ctx context.Context, appID string, start time.Time, end time.Time) ([]model.UsageCreditsUsedByDay, error) {
	rows, err := c.Q.GetUsageCreditsUsedByDayBetween(ctx, pgmodel.GetUsageCreditsUsedByDayBetweenParams{
		AppID:   appID,
		StartAt: pgtype.Timestamp{Time: start, Valid: true},
		EndAt:   pgtype.Timestamp{Time: end, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	records := make([]model.UsageCreditsUsedByDay, 0, len(rows))
	for _, row := range rows {
		records = append(records, model.UsageCreditsUsedByDay{
			Date:        row.Date.Time,
			CreditsUsed: int(row.CreditsUsed),
		})
	}

	return records, nil
}

func (c *Client) AllUsageCreditsUsedBetween(ctx context.Context, start time.Time, end time.Time) (map[string]int, error) {
	rows, err := c.Q.GetAllUsageCreditsUsedBetween(ctx, pgmodel.GetAllUsageCreditsUsedBetweenParams{
		StartAt: pgtype.Timestamp{Time: start, Valid: true},
		EndAt:   pgtype.Timestamp{Time: end, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	res := make(map[string]int, len(rows))
	for _, row := range rows {
		res[row.AppID] = int(row.Sum)
	}

	return res, nil
}

func rowToUsageRecord(row pgmodel.UsageRecord) model.UsageRecord {
	return model.UsageRecord{
		ID:              row.ID,
		Type:            model.UsageRecordType(row.Type),
		AppID:           row.AppID,
		CommandID:       null.NewString(row.CommandID.String, row.CommandID.Valid),
		EventListenerID: null.NewString(row.EventListenerID.String, row.EventListenerID.Valid),
		MessageID:       null.NewString(row.MessageID.String, row.MessageID.Valid),
		CreditsUsed:     int(row.CreditsUsed),
		CreatedAt:       row.CreatedAt.Time,
	}
}
