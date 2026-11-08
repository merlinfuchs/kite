package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func (c *Client) CreateNotification(ctx context.Context, n *model.Notification) (bool, error) {
	data := n.Data
	if data == nil {
		data = []byte("{}")
	}

	rows, err := c.Q.CreateNotification(ctx, pgmodel.CreateNotificationParams{
		ID:        n.ID,
		AppID:     pgtype.Text{String: n.AppID.String, Valid: n.AppID.Valid},
		UserID:    pgtype.Text{String: n.UserID.String, Valid: n.UserID.Valid},
		Type:      string(n.Type),
		DedupeKey: pgtype.Text{String: n.DedupeKey.String, Valid: n.DedupeKey.Valid},
		Data:      data,
		CreatedAt: pgtype.Timestamp{Time: n.CreatedAt.UTC(), Valid: true},
	})
	return rows > 0, err
}

func (c *Client) DeleteNotificationsBefore(ctx context.Context, createdBefore time.Time, batchSize int) (int64, error) {
	return c.Q.DeleteNotificationsBefore(ctx, pgmodel.DeleteNotificationsBeforeParams{
		CreatedBefore: pgtype.Timestamp{Time: createdBefore.UTC(), Valid: true},
		BatchSize:     int32(batchSize),
	})
}
