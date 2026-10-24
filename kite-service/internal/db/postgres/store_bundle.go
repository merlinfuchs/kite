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

func (c *Client) Bundle(ctx context.Context, id string) (*model.Bundle, error) {
	row, err := c.Q.GetBundleCurrentVersion(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	var listing *pgmodel.BundleListing
	if row.ListingIcon.Valid {
		listing = &pgmodel.BundleListing{
			BundleID:  row.Bundle.ID,
			Icon:      row.ListingIcon.String,
			Category:  row.ListingCategory.String,
			About:     row.ListingAbout,
			Faq:       row.ListingFaq,
			Setup:     row.ListingSetup,
			Position:  row.ListingPosition.Int32,
			UpdatedAt: row.ListingUpdatedAt,
		}
	}
	b, err := rowToBundle(row.Bundle, listing, row.Content)
	if err != nil {
		return nil, err
	}

	noteRows, err := c.Q.GetBundleVersionNotes(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, n := range noteRows {
		b.VersionNotes = append(b.VersionNotes, model.BundleVersionNotes{
			Version:   int(n.Version),
			Notes:     n.Notes,
			CreatedAt: n.CreatedAt.Time,
		})
	}
	return b, nil
}

func (c *Client) OfficialBundles(ctx context.Context) ([]*model.Bundle, error) {
	rows, err := c.Q.GetOfficialBundlesCurrentVersion(ctx)
	if err != nil {
		return nil, err
	}

	noteRows, err := c.Q.GetOfficialBundleVersionNotes(ctx)
	if err != nil {
		return nil, err
	}
	notes := map[string][]model.BundleVersionNotes{}
	for _, n := range noteRows {
		notes[n.BundleID] = append(notes[n.BundleID], model.BundleVersionNotes{
			Version:   int(n.Version),
			Notes:     n.Notes,
			CreatedAt: n.CreatedAt.Time,
		})
	}

	bundles := make([]*model.Bundle, len(rows))
	for i, row := range rows {
		if bundles[i], err = rowToBundle(row.Bundle, &row.BundleListing, row.Content); err != nil {
			return nil, err
		}
		bundles[i].VersionNotes = notes[row.Bundle.ID]
	}
	return bundles, nil
}

func (c *Client) UpsertOfficialBundle(ctx context.Context, bundle *model.Bundle) error {
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
		BundleID:  bundle.ID,
		Version:   int32(bundle.Version),
		Content:   bundle.Content,
		Notes:     bundle.Notes,
		CreatedAt: pgtype.Timestamp{Time: bundle.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert bundle version: %w", err)
	}

	// Notes of earlier versions can be written after them.
	for _, n := range bundle.VersionNotes {
		if n.Version == bundle.Version {
			continue
		}
		err := q.SetBundleVersionNotes(ctx, pgmodel.SetBundleVersionNotesParams{
			BundleID: bundle.ID,
			Version:  int32(n.Version),
			Notes:    n.Notes,
		})
		if err != nil {
			return fmt.Errorf("failed to set bundle version notes: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (c *Client) BundlesByOwner(ctx context.Context, userID string) ([]*model.Bundle, error) {
	rows, err := c.Q.GetBundlesByOwner(ctx, pgtype.Text{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	bundles := make([]*model.Bundle, len(rows))
	for i, row := range rows {
		if bundles[i], err = rowToBundle(row.Bundle, nil, nil); err != nil {
			return nil, err
		}
		bundles[i].ItemCount = int(row.ItemCount)
	}
	return bundles, nil
}

func (c *Client) CreateBundle(ctx context.Context, bundle *model.Bundle, maxPerOwner int) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	if err := q.LockBundleOwner(ctx, bundle.OwnerUserID.String); err != nil {
		return fmt.Errorf("failed to lock bundle owner: %w", err)
	}
	count, err := q.CountBundlesByOwner(ctx, pgtype.Text{String: bundle.OwnerUserID.String, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to count bundles: %w", err)
	}
	if int(count) >= maxPerOwner {
		return store.ErrLimitReached
	}

	err = q.CreateBundle(ctx, pgmodel.CreateBundleParams{
		ID:          bundle.ID,
		Name:        bundle.Name,
		Description: bundle.Description,
		OwnerUserID: pgtype.Text{String: bundle.OwnerUserID.String, Valid: bundle.OwnerUserID.Valid},
		SourceAppID: pgtype.Text{String: bundle.SourceAppID.String, Valid: bundle.SourceAppID.Valid},
		Version:     int32(bundle.Version),
		CreatedAt:   pgtype.Timestamp{Time: bundle.CreatedAt.UTC(), Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: bundle.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create bundle: %w", err)
	}

	if err := createBundleVersion(ctx, q, bundle); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// keptBundleVersions is how many versions of a user's bundle are kept. Only
// the current one is installed, the others are history.
const keptBundleVersions = 10

func (c *Client) PublishBundleVersion(ctx context.Context, bundle *model.Bundle, fromVersion int) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	rows, err := q.UpdateBundleVersion(ctx, pgmodel.UpdateBundleVersionParams{
		ID:          bundle.ID,
		Name:        bundle.Name,
		Description: bundle.Description,
		Version:     int32(bundle.Version),
		UpdatedAt:   pgtype.Timestamp{Time: bundle.UpdatedAt.UTC(), Valid: true},
		FromVersion: int32(fromVersion),
	})
	if err != nil {
		return fmt.Errorf("failed to update bundle: %w", err)
	}
	if rows == 0 {
		return store.ErrOutdated
	}

	if err := createBundleVersion(ctx, q, bundle); err != nil {
		return err
	}
	err = q.DeleteOldBundleVersions(ctx, pgmodel.DeleteOldBundleVersionsParams{
		BundleID:      bundle.ID,
		BeforeVersion: int32(bundle.Version - keptBundleVersions),
	})
	if err != nil {
		return fmt.Errorf("failed to delete old bundle versions: %w", err)
	}
	return tx.Commit(ctx)
}

func createBundleVersion(ctx context.Context, q *pgmodel.Queries, bundle *model.Bundle) error {
	err := q.CreateBundleVersion(ctx, pgmodel.CreateBundleVersionParams{
		BundleID:  bundle.ID,
		Version:   int32(bundle.Version),
		Content:   bundle.Content,
		Notes:     bundle.Notes,
		CreatedAt: pgtype.Timestamp{Time: bundle.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create bundle version: %w", err)
	}
	return nil
}

func (c *Client) SetBundlePublishedAt(ctx context.Context, id string, publishedAt null.Time) error {
	return c.Q.SetBundlePublishedAt(ctx, pgmodel.SetBundlePublishedAtParams{
		ID:          id,
		PublishedAt: pgtype.Timestamp{Time: publishedAt.Time.UTC(), Valid: publishedAt.Valid},
	})
}

func (c *Client) DeleteBundle(ctx context.Context, id string) error {
	return c.Q.DeleteBundle(ctx, id)
}

func upsertBundleListing(ctx context.Context, q *pgmodel.Queries, bundleID string, listing model.BundleListing) error {
	about, err := marshalList(listing.About)
	if err != nil {
		return err
	}
	faq, err := marshalList(listing.FAQ)
	if err != nil {
		return err
	}
	setup, err := marshalList(listing.Setup)
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
		Position:  int32(listing.Position),
		UpdatedAt: pgtype.Timestamp{Time: listing.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert bundle listing: %w", err)
	}
	return nil
}

// marshalList encodes a nil list as [] instead of null.
func marshalList[T any](list []T) ([]byte, error) {
	if list == nil {
		list = []T{}
	}
	return json.Marshal(list)
}

func rowToBundle(row pgmodel.Bundle, listing *pgmodel.BundleListing, content []byte) (*model.Bundle, error) {
	bundle := &model.Bundle{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		OwnerUserID: null.NewString(row.OwnerUserID.String, row.OwnerUserID.Valid),
		SourceAppID: null.NewString(row.SourceAppID.String, row.SourceAppID.Valid),
		PublishedAt: null.NewTime(row.PublishedAt.Time, row.PublishedAt.Valid),
		Version:     int(row.Version),
		Content:     content,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
	if listing != nil {
		var err error
		if bundle.Listing, err = rowToBundleListing(*listing); err != nil {
			return nil, err
		}
	}
	return bundle, nil
}

func rowToBundleListing(row pgmodel.BundleListing) (*model.BundleListing, error) {
	listing := &model.BundleListing{
		Icon:      row.Icon,
		Category:  row.Category,
		Position:  int(row.Position),
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
