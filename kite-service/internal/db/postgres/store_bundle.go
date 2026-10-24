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

func (c *Client) BundleCurrentVersion(ctx context.Context, id string) (*model.Bundle, *model.BundleVersion, error) {
	row, err := c.Q.GetBundleCurrentVersion(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, store.ErrNotFound
		}
		return nil, nil, err
	}

	return rowToBundleWithVersion(row.Bundle, nil, row.Content)
}

func (c *Client) ListedBundleCurrentVersion(ctx context.Context, id string) (*model.Bundle, *model.BundleVersion, error) {
	row, err := c.Q.GetListedBundleCurrentVersion(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, store.ErrNotFound
		}
		return nil, nil, err
	}

	return rowToBundleWithVersion(row.Bundle, &row.BundleListing, row.Content)
}

func (c *Client) OfficialBundlesCurrentVersion(ctx context.Context) ([]*model.Bundle, []*model.BundleVersion, error) {
	rows, err := c.Q.GetOfficialBundlesCurrentVersion(ctx)
	if err != nil {
		return nil, nil, err
	}

	bundles := make([]*model.Bundle, len(rows))
	versions := make([]*model.BundleVersion, len(rows))
	for i, row := range rows {
		if bundles[i], versions[i], err = rowToBundleWithVersion(row.Bundle, &row.BundleListing, row.Content); err != nil {
			return nil, nil, err
		}
	}
	return bundles, versions, nil
}

func (c *Client) UpsertOfficialBundle(ctx context.Context, bundle *model.Bundle, version *model.BundleVersion) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	rows, err := q.UpsertOfficialBundle(ctx, pgmodel.UpsertOfficialBundleParams{
		ID:          bundle.ID,
		Name:        bundle.Name,
		Description: bundle.Description,
		Version:     int32(bundle.Version),
		CreatedAt:   pgtype.Timestamp{Time: bundle.CreatedAt.UTC(), Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: bundle.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert bundle: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bundle %s belongs to a user", bundle.ID)
	}

	if bundle.Listing != nil {
		if err := upsertBundleListing(ctx, q, bundle.ID, *bundle.Listing); err != nil {
			return err
		}
	}

	err = q.UpsertBundleVersion(ctx, pgmodel.UpsertBundleVersionParams{
		BundleID:  version.BundleID,
		Version:   int32(version.Version),
		Content:   version.Content,
		CreatedAt: pgtype.Timestamp{Time: version.CreatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert bundle version: %w", err)
	}

	return tx.Commit(ctx)
}

func upsertBundleListing(ctx context.Context, q *pgmodel.Queries, bundleID string, listing model.BundleListing) error {
	about, err := json.Marshal(listing.About)
	if err != nil {
		return err
	}
	faq, err := json.Marshal(listing.FAQ)
	if err != nil {
		return err
	}
	setup, err := json.Marshal(listing.Setup)
	if err != nil {
		return err
	}

	err = q.UpsertBundleListing(ctx, pgmodel.UpsertBundleListingParams{
		BundleID:  bundleID,
		Icon:      listing.Icon,
		Category:  listing.Category,
		About:     about,
		Faq:       faq,
		Setup:     setup,
		UpdatedAt: pgtype.Timestamp{Time: listing.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert bundle listing: %w", err)
	}
	return nil
}

func rowToBundleWithVersion(row pgmodel.Bundle, listing *pgmodel.BundleListing, content []byte) (*model.Bundle, *model.BundleVersion, error) {
	bundle := &model.Bundle{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		OwnerUserID: null.NewString(row.OwnerUserID.String, row.OwnerUserID.Valid),
		Version:     int(row.Version),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
	if listing != nil {
		var err error
		if bundle.Listing, err = rowToBundleListing(*listing); err != nil {
			return nil, nil, err
		}
	}

	return bundle, &model.BundleVersion{
		BundleID: bundle.ID,
		Version:  bundle.Version,
		Content:  content,
	}, nil
}

func rowToBundleListing(row pgmodel.BundleListing) (*model.BundleListing, error) {
	listing := &model.BundleListing{
		Icon:      row.Icon,
		Category:  row.Category,
		UpdatedAt: row.UpdatedAt.Time,
	}
	if err := json.Unmarshal(row.About, &listing.About); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle about: %w", err)
	}
	if err := json.Unmarshal(row.Faq, &listing.FAQ); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle faq: %w", err)
	}
	if err := json.Unmarshal(row.Setup, &listing.Setup); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle setup: %w", err)
	}
	return listing, nil
}
