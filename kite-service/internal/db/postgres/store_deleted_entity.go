package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func (c *Client) DeletedEntitiesSince(ctx context.Context, deletedSince time.Time) ([]*model.DeletedEntity, error) {
	rows, err := c.Q.GetDeletedEntitiesSince(ctx, pgtype.Timestamp{
		Time:  deletedSince.UTC(),
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	entities := make([]*model.DeletedEntity, len(rows))
	for i, row := range rows {
		entities[i] = &model.DeletedEntity{
			ID:        row.ID,
			Type:      model.DeletedEntityType(row.EntityType),
			AppID:     row.AppID,
			DeletedAt: row.DeletedAt.Time,
		}
	}

	return entities, nil
}

func (c *Client) DeleteDeletedEntitiesBefore(ctx context.Context, deletedBefore time.Time) error {
	return c.Q.DeleteDeletedEntitiesBefore(ctx, pgtype.Timestamp{
		Time:  deletedBefore.UTC(),
		Valid: true,
	})
}
