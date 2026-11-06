package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func (c *Client) Stats(ctx context.Context, executionsSince time.Time) (model.Stats, error) {
	row, err := c.Q.GetStats(ctx, pgtype.Timestamp{Time: executionsSince, Valid: true})
	if err != nil {
		return model.Stats{}, err
	}

	return model.Stats{
		Users:          int(row.Users),
		Apps:           int(row.Apps),
		Commands:       int(row.Commands),
		EventListeners: int(row.EventListeners),
		Executions:     int(row.Executions),
	}, nil
}
