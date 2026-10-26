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
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	bundle, err := q.LockBundle(ctx, review.BundleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}
		return fmt.Errorf("failed to lock bundle: %w", err)
	}
	if bundle.RemovedAt.Valid {
		return store.ErrNotFound
	}

	review.Version = int(bundle.Version)
	if err := submitBundleReview(ctx, q, review); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func submitBundleReview(ctx context.Context, q *pgmodel.Queries, review *model.BundleReview) error {
	// Lists that are left out are empty, like in bundle_listings.
	l := review.Listing
	l.About = emptyIfNil(l.About)
	l.FAQ = emptyIfNil(l.FAQ)
	l.Setup = emptyIfNil(l.Setup)
	listing, err := json.Marshal(l)
	if err != nil {
		return fmt.Errorf("failed to encode bundle listing: %w", err)
	}
	return q.UpsertPendingBundleReview(ctx, pgmodel.UpsertPendingBundleReviewParams{
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
		b.ListedVersion = int(row.ListedVersion.Int32)
		if row.Notes != "" {
			b.VersionNotes = []model.BundleVersionNotes{{Version: review.Version, Notes: row.Notes}}
		}
		b.Owner = rowToUser(row.User)
		res[i] = &model.PendingBundleReview{Review: review, Bundle: b}
	}
	return res, nil
}

func (c *Client) DecideBundleReview(ctx context.Context, decision *model.BundleReview) (*model.BundleReview, error) {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	bundleID, err := q.GetPendingBundleReviewBundleID(ctx, decision.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get bundle review: %w", err)
	}
	// A bundle that was taken down has no pending review anymore.
	bundle, err := q.LockBundle(ctx, bundleID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock bundle: %w", err)
	}
	if bundle.RemovedAt.Valid {
		return nil, store.ErrNotFound
	}

	row, err := q.GetPendingBundleReviewForUpdate(ctx, decision.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get bundle review: %w", err)
	}
	review, err := rowToBundleReview(row)
	if err != nil {
		return nil, err
	}
	review.Status = decision.Status
	review.Reason = decision.Reason
	review.ReviewerUserID = decision.ReviewerUserID
	review.DecidedAt = decision.DecidedAt

	err = q.DecideBundleReview(ctx, pgmodel.DecideBundleReviewParams{
		ID:             review.ID,
		Status:         string(review.Status),
		Reason:         review.Reason,
		ReviewerUserID: pgtype.Text{String: review.ReviewerUserID.String, Valid: review.ReviewerUserID.Valid},
		DecidedAt:      pgtype.Timestamp{Time: review.DecidedAt.Time.UTC(), Valid: review.DecidedAt.Valid},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decide bundle review: %w", err)
	}

	if review.Status == model.BundleReviewStatusApproved {
		listing := review.Listing
		listing.UpdatedAt = review.DecidedAt.Time
		if err := upsertBundleListing(ctx, q, review.BundleID, listing, review.Version); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return review, nil
}

func emptyIfNil[T any](list []T) []T {
	if list == nil {
		return []T{}
	}
	return list
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
