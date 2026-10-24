package postgres

import (
	"context"
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

	bundle := rowToBundle(row.Bundle)
	return bundle, &model.BundleVersion{
		BundleID: bundle.ID,
		Version:  bundle.Version,
		Content:  row.Content,
	}, nil
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

func rowToBundle(row pgmodel.Bundle) *model.Bundle {
	return &model.Bundle{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		OwnerUserID: null.NewString(row.OwnerUserID.String, row.OwnerUserID.Valid),
		Version:     int(row.Version),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
