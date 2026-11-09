package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreateAppAuditLogEntry(ctx context.Context, entry *model.AppAuditLogEntry) error {
	return c.Q.CreateAppAuditLog(ctx, pgmodel.CreateAppAuditLogParams{
		AppID:      entry.AppID,
		UserID:     pgtype.Text{String: entry.UserID.String, Valid: entry.UserID.Valid},
		Action:     entry.Action,
		TargetType: entry.TargetType,
		TargetID:   entry.TargetID,
		TargetName: entry.TargetName,
		ApiTokenID: pgtype.Text{String: entry.APITokenID.String, Valid: entry.APITokenID.Valid},
		CreatedAt:  pgtype.Timestamp{Time: entry.CreatedAt.UTC(), Valid: true},
	})
}

func (c *Client) AppAuditLogEntries(ctx context.Context, appID string, beforeID int64, limit int) ([]*model.AppAuditLogEntry, error) {
	rows, err := c.Q.GetAppAuditLogs(ctx, pgmodel.GetAppAuditLogsParams{
		AppID:      appID,
		BeforeID:   beforeID,
		LimitCount: int32(limit),
	})
	if err != nil {
		return nil, err
	}

	entries := make([]*model.AppAuditLogEntry, len(rows))
	for i, row := range rows {
		entry := &model.AppAuditLogEntry{
			ID:         row.ID,
			AppID:      row.AppID,
			UserID:     null.NewString(row.UserID.String, row.UserID.Valid),
			Action:     row.Action,
			TargetType: row.TargetType,
			TargetID:   row.TargetID,
			TargetName: row.TargetName,
			APITokenID: null.NewString(row.ApiTokenID.String, row.ApiTokenID.Valid),
			CreatedAt:  row.CreatedAt.Time,
		}
		if row.UserID.Valid && row.UserDisplayName.Valid {
			entry.User = &model.User{
				ID:              row.UserID.String,
				DisplayName:     row.UserDisplayName.String,
				DiscordID:       row.UserDiscordID.String,
				DiscordUsername: row.UserDiscordUsername.String,
				DiscordAvatar:   null.NewString(row.UserDiscordAvatar.String, row.UserDiscordAvatar.Valid),
			}
		}
		entries[i] = entry
	}

	return entries, nil
}

func (c *Client) DeleteAppAuditLogEntriesBefore(ctx context.Context, before time.Time) error {
	return c.Q.DeleteAppAuditLogsBefore(ctx, pgtype.Timestamp{Time: before.UTC(), Valid: true})
}
