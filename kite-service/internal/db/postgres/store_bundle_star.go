package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
)

func (c *Client) StarBundle(ctx context.Context, bundleID string, userID string, starred bool, at time.Time) (int, error) {
	var err error
	if starred {
		err = c.Q.CreateBundleStar(ctx, pgmodel.CreateBundleStarParams{
			BundleID:  bundleID,
			UserID:    userID,
			CreatedAt: pgtype.Timestamp{Time: at.UTC(), Valid: true},
		})
	} else {
		err = c.Q.DeleteBundleStar(ctx, pgmodel.DeleteBundleStarParams{BundleID: bundleID, UserID: userID})
	}
	if err != nil {
		return 0, err
	}
	count, err := c.Q.CountBundleStars(ctx, bundleID)
	return int(count), err
}

func (c *Client) StarredBundleIDs(ctx context.Context, userID string) ([]string, error) {
	return c.Q.GetStarredBundleIDs(ctx, userID)
}
