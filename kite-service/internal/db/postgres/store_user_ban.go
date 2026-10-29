package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) UserBans(ctx context.Context) ([]*model.UserBan, error) {
	rows, err := c.Q.GetUserBans(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*model.UserBan, len(rows))
	for i, row := range rows {
		res[i] = rowToUserBan(row.UserBan)
		res[i].User = rowToUser(row.User)
	}
	return res, nil
}

func (c *Client) UserBan(ctx context.Context, userID string, feature model.UserBanFeature) (*model.UserBan, error) {
	row, err := c.Q.GetUserBan(ctx, pgmodel.GetUserBanParams{
		UserID:  userID,
		Feature: string(feature),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return rowToUserBan(row), nil
}

// marketplaceBan returns the user's ban from the marketplace, or nil if they
// aren't banned from it. Users banned from Kite can't publish versions at all.
func marketplaceBan(ctx context.Context, q *pgmodel.Queries, userID string) (*model.UserBan, error) {
	row, err := q.GetUserBan(ctx, pgmodel.GetUserBanParams{
		UserID:  userID,
		Feature: string(model.UserBanFeatureMarketplace),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user ban: %w", err)
	}
	return rowToUserBan(row), nil
}

func (c *Client) BanUser(ctx context.Context, ban *model.UserBan) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	err = q.UpsertUserBan(ctx, pgmodel.UpsertUserBanParams{
		UserID:         ban.UserID,
		Feature:        string(ban.Feature),
		Reason:         ban.Reason,
		BannedByUserID: pgtype.Text{String: ban.BannedByUserID.String, Valid: ban.BannedByUserID.Valid},
		CreatedAt:      pgtype.Timestamp{Time: ban.CreatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert user ban: %w", err)
	}

	// A ban from Kite includes the marketplace.
	if ban.Feature == model.UserBanFeatureMarketplace || ban.Feature == model.UserBanFeatureKite {
		err := q.RejectPendingBundleReviewsByOwner(ctx, pgmodel.RejectPendingBundleReviewsByOwnerParams{
			Reason:         ban.Reason,
			ReviewerUserID: pgtype.Text{String: ban.BannedByUserID.String, Valid: ban.BannedByUserID.Valid},
			DecidedAt:      pgtype.Timestamp{Time: ban.CreatedAt.UTC(), Valid: true},
			OwnerUserID:    pgtype.Text{String: ban.UserID, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to reject pending bundle reviews: %w", err)
		}
	}
	if ban.Feature == model.UserBanFeatureKite {
		err := q.DisableAppsOfOwner(ctx, pgmodel.DisableAppsOfOwnerParams{
			DisabledReason: pgtype.Text{String: model.AppDisabledReasonOwnerBanned, Valid: true},
			UpdatedAt:      pgtype.Timestamp{Time: ban.CreatedAt.UTC(), Valid: true},
			OwnerUserID:    ban.UserID,
		})
		if err != nil {
			return fmt.Errorf("failed to disable apps: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (c *Client) UnbanUser(ctx context.Context, userID string, feature model.UserBanFeature) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	rows, err := q.DeleteUserBan(ctx, pgmodel.DeleteUserBanParams{
		UserID:  userID,
		Feature: string(feature),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrNotFound
	}

	if feature == model.UserBanFeatureKite {
		err := q.EnableAppsOfOwnerDisabledWithReason(ctx, pgmodel.EnableAppsOfOwnerDisabledWithReasonParams{
			UpdatedAt:      pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
			OwnerUserID:    userID,
			DisabledReason: pgtype.Text{String: model.AppDisabledReasonOwnerBanned, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to enable apps: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func rowToUserBan(row pgmodel.UserBan) *model.UserBan {
	return &model.UserBan{
		UserID:         row.UserID,
		Feature:        model.UserBanFeature(row.Feature),
		Reason:         row.Reason,
		BannedByUserID: null.NewString(row.BannedByUserID.String, row.BannedByUserID.Valid),
		CreatedAt:      row.CreatedAt.Time,
	}
}
