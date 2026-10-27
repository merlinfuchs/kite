package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
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
			BundleID:    row.Bundle.ID,
			Name:        row.ListingName.String,
			Description: row.ListingDescription.String,
			Icon:        row.ListingIcon.String,
			Category:    row.ListingCategory.String,
			About:       row.ListingAbout,
			Faq:         row.ListingFaq,
			Setup:       row.ListingSetup,
			Position:    row.ListingPosition.Int32,
			UpdatedAt:   row.ListingUpdatedAt,
		}
	}
	b, err := rowToBundle(row.Bundle, listing, row.Content)
	if err != nil {
		return nil, err
	}
	b.InReview = row.InReview

	b.Version = b.InstallVersion()

	noteRows, err := c.Q.GetBundleVersionNotes(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, n := range noteRows {
		if int(n.Version) > b.Version {
			continue
		}
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

func (c *Client) BundleSummaries(ctx context.Context, ids []string) ([]*model.Bundle, error) {
	rows, err := c.Q.GetBundleSummaries(ctx, ids)
	if err != nil {
		return nil, err
	}

	bundles := make([]*model.Bundle, len(rows))
	for i, row := range rows {
		b, err := rowToBundle(row.Bundle, nil, nil)
		if err != nil {
			return nil, err
		}
		// Lists only show the icon of the listing.
		if row.ListingIcon.Valid {
			b.Listing = &model.BundleListing{Icon: row.ListingIcon.String}
		}
		if row.OwnerDisplayName.Valid {
			b.Owner = &model.User{
				ID:              b.OwnerUserID.String,
				DisplayName:     row.OwnerDisplayName.String,
				DiscordID:       row.OwnerDiscordID.String,
				DiscordUsername: row.OwnerDiscordUsername.String,
				DiscordAvatar:   null.NewString(row.OwnerDiscordAvatar.String, row.OwnerDiscordAvatar.Valid),
			}
		}
		bundles[i] = b
	}
	return bundles, nil
}

func (c *Client) ListedUserBundles(ctx context.Context) ([]*model.Bundle, error) {
	rows, err := c.Q.GetListedUserBundles(ctx)
	if err != nil {
		return nil, err
	}

	bundles := make([]*model.Bundle, len(rows))
	for i, row := range rows {
		b, err := rowToBundle(row.Bundle, &row.BundleListing, row.Content)
		if err != nil {
			return nil, err
		}
		b.Version = b.InstallVersion()
		b.Owner = rowToUser(row.User)
		bundles[i] = b
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

	reviewRows, err := c.Q.GetLatestBundleReviewsByOwner(ctx, pgtype.Text{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	reviews := make(map[string]*model.BundleReview, len(reviewRows))
	for _, r := range reviewRows {
		if reviews[r.BundleID], err = rowToBundleReview(r); err != nil {
			return nil, err
		}
	}

	bundles := make([]*model.Bundle, len(rows))
	for i, row := range rows {
		if bundles[i], err = rowToBundle(row.Bundle, nil, nil); err != nil {
			return nil, err
		}
		bundles[i].ItemCount = int(row.ItemCount)
		bundles[i].InstallCount = int(row.InstallCount)
		bundles[i].LatestReview = reviews[row.Bundle.ID]
		bundles[i].Listed = row.Listed
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

func (c *Client) PublishBundleVersion(ctx context.Context, bundle *model.Bundle, fromVersion int) (bool, error) {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	row, err := q.UpdateBundleVersion(ctx, pgmodel.UpdateBundleVersionParams{
		ID:          bundle.ID,
		Version:     int32(bundle.Version),
		UpdatedAt:   pgtype.Timestamp{Time: bundle.UpdatedAt.UTC(), Valid: true},
		FromVersion: int32(fromVersion),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, store.ErrOutdated
	}
	if err != nil {
		return false, fmt.Errorf("failed to update bundle: %w", err)
	}

	if err := createBundleVersion(ctx, q, bundle); err != nil {
		return false, err
	}

	submitted, err := reviewNewVersion(ctx, q, bundle, row)
	if err != nil {
		return false, err
	}

	err = q.DeleteOldBundleVersions(ctx, pgmodel.DeleteOldBundleVersionsParams{
		BundleID:      bundle.ID,
		BeforeVersion: int32(bundle.Version - keptBundleVersions),
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete old bundle versions: %w", err)
	}
	return submitted, tx.Commit(ctx)
}

// reviewNewVersion makes a pending review review the new version instead,
// and returns whether it submitted a new review. A listed bundle without
// one gets one with its page, which apps only get the new version after. So
// does a rejected bundle, so publishing a fix is enough to get it reviewed
// again, unless its owner is banned from the marketplace. The row is the
// bundle as its version was updated.
func reviewNewVersion(ctx context.Context, q *pgmodel.Queries, bundle *model.Bundle, row pgmodel.Bundle) (bool, error) {
	if bundle.RemovedAt.Valid {
		return false, nil
	}

	moved, err := q.SetPendingBundleReviewVersion(ctx, pgmodel.SetPendingBundleReviewVersionParams{
		ID:       util.UniqueID(),
		BundleID: bundle.ID,
		Version:  int32(bundle.Version),
	})
	if err != nil {
		return false, fmt.Errorf("failed to update pending bundle review: %w", err)
	}
	if moved > 0 {
		return false, nil
	}

	resubmit, err := newVersionsReviewed(ctx, q, bundle.ID)
	if err != nil || !resubmit {
		return false, err
	}
	ban, err := marketplaceBan(ctx, q, row.OwnerUserID.String)
	if err != nil || ban != nil {
		return false, err
	}
	page, err := rowToBundlePage(row)
	if err != nil {
		return false, err
	}
	// The page of a rejected bundle can change without a review, and the
	// marketplace needs it complete.
	if !page.Complete() {
		return false, nil
	}
	err = submitBundleReview(ctx, q, &model.BundleReview{
		ID:        util.UniqueID(),
		BundleID:  bundle.ID,
		Version:   bundle.Version,
		Listing:   page,
		CreatedAt: bundle.UpdatedAt,
	})
	return err == nil, err
}

// newVersionsReviewed is whether a new version of the bundle is reviewed:
// while it's listed, or when its last review rejected it.
func newVersionsReviewed(ctx context.Context, q *pgmodel.Queries, bundleID string) (bool, error) {
	_, err := q.GetBundleListing(ctx, bundleID)
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("failed to get bundle listing: %w", err)
	}

	review, err := q.GetLatestBundleReview(ctx, bundleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get latest bundle review: %w", err)
	}
	return review.Status == string(model.BundleReviewStatusRejected), nil
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

func (c *Client) UpdateBundlePage(ctx context.Context, id string, page model.BundleListing) error {
	return updateBundlePage(ctx, c.Q, id, page)
}

func updateBundlePage(ctx context.Context, q *pgmodel.Queries, id string, page model.BundleListing) error {
	about, err := marshalList(page.About)
	if err != nil {
		return err
	}
	setup, err := marshalList(page.Setup)
	if err != nil {
		return err
	}
	rows, err := q.UpdateBundlePage(ctx, pgmodel.UpdateBundlePageParams{
		ID:          id,
		Name:        page.Name,
		Description: page.Description,
		Icon:        page.Icon,
		Category:    page.Category,
		About:       about,
		Setup:       setup,
	})
	if err != nil {
		return fmt.Errorf("failed to update bundle page: %w", err)
	}
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (c *Client) UnlistBundle(ctx context.Context, id string) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := unlistBundle(ctx, c.Q.WithTx(tx), id, time.Now().UTC()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// unlistBundle takes the bundle out of the marketplace and ends its pending
// review.
func unlistBundle(ctx context.Context, q *pgmodel.Queries, id string, at time.Time) error {
	if err := q.DeleteBundleListing(ctx, id); err != nil {
		return fmt.Errorf("failed to delete bundle listing: %w", err)
	}
	err := q.WithdrawOpenBundleReviews(ctx, pgmodel.WithdrawOpenBundleReviewsParams{
		BundleID: id,
		At:       pgtype.Timestamp{Time: at.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to withdraw open bundle reviews: %w", err)
	}
	return nil
}

func (c *Client) DeleteBundle(ctx context.Context, id string) error {
	rows, err := c.Q.DeleteBundle(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

// upsertBundleListing stores how the bundle is shown in the marketplace.
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
		BundleID:    bundleID,
		Name:        listing.Name,
		Description: listing.Description,
		Icon:        listing.Icon,
		Category:    listing.Category,
		About:       about,
		Faq:         faq,
		Setup:       setup,
		Position:    int32(listing.Position),
		UpdatedAt:   pgtype.Timestamp{Time: listing.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert bundle listing: %w", err)
	}
	return nil
}

// marshalList encodes a nil list as [] instead of null.
func marshalList[T any](list []T) ([]byte, error) {
	return json.Marshal(emptyIfNil(list))
}

// rowToBundle shows the bundle with the listing if it's listed, otherwise with
// its page.
func rowToBundle(row pgmodel.Bundle, listing *pgmodel.BundleListing, content []byte) (*model.Bundle, error) {
	page, err := rowToBundlePage(row)
	if err != nil {
		return nil, err
	}
	bundle := &model.Bundle{
		ID:            row.ID,
		Name:          row.Name,
		Description:   row.Description,
		OwnerUserID:   null.NewString(row.OwnerUserID.String, row.OwnerUserID.Valid),
		SourceAppID:   null.NewString(row.SourceAppID.String, row.SourceAppID.Valid),
		RemovedAt:     null.NewTime(row.RemovedAt.Time, row.RemovedAt.Valid),
		RemovedReason: row.RemovedReason,
		Version:       int(row.Version),
		LatestVersion: int(row.Version),
		// Official bundles are never reviewed.
		ApprovedVersion: int(row.ApprovedVersion.Int32),
		Content:         content,
		Page:            page,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
	if listing != nil {
		l, err := rowToBundleListing(*listing)
		if err != nil {
			return nil, err
		}
		bundle.ShowListing(l)
		bundle.Listed = true
	} else if bundle.OwnerUserID.Valid {
		bundle.Listing = &page
	}
	return bundle, nil
}

// rowToBundlePage is the page of a user's bundle. Official bundles have only
// their name and description in their row.
func rowToBundlePage(row pgmodel.Bundle) (model.BundleListing, error) {
	page, err := rowToBundleListing(pgmodel.BundleListing{
		Name:        row.Name,
		Description: row.Description,
		Icon:        row.Icon,
		Category:    row.Category,
		About:       row.About,
		Faq:         []byte("[]"),
		Setup:       row.Setup,
	})
	if err != nil {
		return model.BundleListing{}, err
	}
	return *page, nil
}

func rowToBundleListing(row pgmodel.BundleListing) (*model.BundleListing, error) {
	listing := &model.BundleListing{
		Name:        row.Name,
		Description: row.Description,
		Icon:        row.Icon,
		Category:    row.Category,
		Position:    int(row.Position),
		UpdatedAt:   row.UpdatedAt.Time,
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
