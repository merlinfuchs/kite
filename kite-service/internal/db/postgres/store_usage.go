package postgres

import (
	"context"
	"fmt"
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
		GuildID:         pgtype.Text{String: record.GuildID.String, Valid: record.GuildID.Valid},
		UserID:          pgtype.Text{String: record.UserID.String, Valid: record.UserID.Valid},
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

func (c *Client) UsageCreditsUsedByTargetSince(ctx context.Context, appID string, scope model.CreditLimitScope, targetID string, since time.Time) (int, error) {
	startAt := pgtype.Timestamp{Time: since.UTC(), Valid: true}

	var res int32
	var err error
	switch scope {
	case model.CreditLimitScopeGuild:
		res, err = c.Q.GetUsageCreditsUsedByGuildSince(ctx, pgmodel.GetUsageCreditsUsedByGuildSinceParams{
			AppID:   appID,
			GuildID: pgtype.Text{String: targetID, Valid: true},
			StartAt: startAt,
		})
	case model.CreditLimitScopeUser:
		res, err = c.Q.GetUsageCreditsUsedByUserSince(ctx, pgmodel.GetUsageCreditsUsedByUserSinceParams{
			AppID:   appID,
			UserID:  pgtype.Text{String: targetID, Valid: true},
			StartAt: startAt,
		})
	default:
		return 0, fmt.Errorf("unknown credit limit scope: %s", scope)
	}
	if err != nil {
		return 0, err
	}

	return int(res), nil
}

func (c *Client) TopUsageCreditsByTargetBetween(ctx context.Context, appID string, scope model.CreditLimitScope, start time.Time, end time.Time, limit int) ([]model.UsageCreditsUsedByTarget, error) {
	startAt := pgtype.Timestamp{Time: start, Valid: true}
	endAt := pgtype.Timestamp{Time: end, Valid: true}

	var res []model.UsageCreditsUsedByTarget
	switch scope {
	case model.CreditLimitScopeGuild:
		rows, err := c.Q.GetTopUsageCreditsByGuildBetween(ctx, pgmodel.GetTopUsageCreditsByGuildBetweenParams{
			AppID:    appID,
			StartAt:  startAt,
			EndAt:    endAt,
			RowLimit: int32(limit),
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			res = append(res, model.UsageCreditsUsedByTarget{TargetID: row.TargetID, CreditsUsed: int(row.CreditsUsed)})
		}
	case model.CreditLimitScopeUser:
		rows, err := c.Q.GetTopUsageCreditsByUserBetween(ctx, pgmodel.GetTopUsageCreditsByUserBetweenParams{
			AppID:    appID,
			StartAt:  startAt,
			EndAt:    endAt,
			RowLimit: int32(limit),
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			res = append(res, model.UsageCreditsUsedByTarget{TargetID: row.TargetID, CreditsUsed: int(row.CreditsUsed)})
		}
	default:
		return nil, fmt.Errorf("unknown credit limit scope: %s", scope)
	}

	return res, nil
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

func (c *Client) DeleteUsageRecordsBefore(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	return c.Q.DeleteUsageRecordsBefore(ctx, pgmodel.DeleteUsageRecordsBeforeParams{
		BeforeAt:  pgtype.Timestamp{Time: before, Valid: true},
		BatchSize: int32(batchSize),
	})
}

func rowToUsageRecord(row pgmodel.UsageRecord) model.UsageRecord {
	return model.UsageRecord{
		ID:              row.ID,
		Type:            model.UsageRecordType(row.Type),
		AppID:           row.AppID,
		CommandID:       null.NewString(row.CommandID.String, row.CommandID.Valid),
		EventListenerID: null.NewString(row.EventListenerID.String, row.EventListenerID.Valid),
		MessageID:       null.NewString(row.MessageID.String, row.MessageID.Valid),
		GuildID:         null.NewString(row.GuildID.String, row.GuildID.Valid),
		UserID:          null.NewString(row.UserID.String, row.UserID.Valid),
		CreditsUsed:     int(row.CreditsUsed),
		CreatedAt:       row.CreatedAt.Time,
	}
}
