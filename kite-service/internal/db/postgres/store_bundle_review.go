package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) SubmitBundleReview(ctx context.Context, review *model.BundleReview) error {
	listing, err := json.Marshal(review.Listing)
	if err != nil {
		return fmt.Errorf("failed to encode bundle listing: %w", err)
	}
	return c.Q.UpsertPendingBundleReview(ctx, pgmodel.UpsertPendingBundleReviewParams{
		ID:        review.ID,
		BundleID:  review.BundleID,
		Version:   int32(review.Version),
		Listing:   listing,
		CreatedAt: pgtype.Timestamp{Time: review.CreatedAt.UTC(), Valid: true},
	})
}

func (c *Client) PendingBundleReviews(ctx context.Context) ([]*model.PendingBundleReview, error) {
	rows, err := c.Q.GetPendingBundleReviews(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*model.PendingBundleReview, len(rows))
	for i, row := range rows {
		review, err := rowToBundleReview(row.BundleReview)
		if err != nil {
			return nil, err
		}
		b, err := rowToBundle(row.Bundle, nil, row.Content)
		if err != nil {
			return nil, err
		}
		b.Version = review.Version
		if row.ListedVersion.Valid {
			b.Listing = &model.BundleListing{ListedVersion: int(row.ListedVersion.Int32)}
		}
		res[i] = &model.PendingBundleReview{Review: review, Bundle: b}
	}
	return res, nil
}

func (c *Client) DecideBundleReview(ctx context.Context, review *model.BundleReview) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	row, err := q.GetPendingBundleReviewForUpdate(ctx, review.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}
		return fmt.Errorf("failed to get bundle review: %w", err)
	}
	pending, err := rowToBundleReview(row)
	if err != nil {
		return err
	}

	err = q.DecideBundleReview(ctx, pgmodel.DecideBundleReviewParams{
		ID:             review.ID,
		Status:         string(review.Status),
		Reason:         review.Reason,
		ReviewerUserID: pgtype.Text{String: review.ReviewerUserID.String, Valid: review.ReviewerUserID.Valid},
		DecidedAt:      pgtype.Timestamp{Time: review.DecidedAt.Time.UTC(), Valid: review.DecidedAt.Valid},
	})
	if err != nil {
		return fmt.Errorf("failed to decide bundle review: %w", err)
	}

	if review.Status == model.BundleReviewStatusApproved {
		listing := pending.Listing
		listing.ListedVersion = pending.Version
		listing.UpdatedAt = review.DecidedAt.Time
		if err := upsertBundleListing(ctx, q, pending.BundleID, listing); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// The caller gets the rest of the review, like its bundle.
	decided := *pending
	decided.Status = review.Status
	decided.Reason = review.Reason
	decided.ReviewerUserID = review.ReviewerUserID
	decided.DecidedAt = review.DecidedAt
	*review = decided
	return nil
}

func rowToBundleReview(row pgmodel.BundleReview) (*model.BundleReview, error) {
	review := &model.BundleReview{
		ID:             row.ID,
		BundleID:       row.BundleID,
		Version:        int(row.Version),
		Status:         model.BundleReviewStatus(row.Status),
		Reason:         row.Reason,
		ReviewerUserID: null.NewString(row.ReviewerUserID.String, row.ReviewerUserID.Valid),
		CreatedAt:      row.CreatedAt.Time,
		DecidedAt:      null.NewTime(row.DecidedAt.Time, row.DecidedAt.Valid),
	}
	if err := json.Unmarshal(row.Listing, &review.Listing); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle review listing: %w", err)
	}
	return review, nil
}
