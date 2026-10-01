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
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreateMarketplaceListing(ctx context.Context, listing *model.MarketplaceListing) (*model.MarketplaceListing, error) {
	items, err := json.Marshal(listing.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal items: %w", err)
	}

	row, err := c.Q.CreateMarketplaceListing(ctx, pgmodel.CreateMarketplaceListingParams{
		ID:                 listing.ID,
		Name:               listing.Name,
		Description:        listing.Description,
		AuthorUserID:       listing.AuthorUserID,
		SourceAppID:        pgtype.Text{String: listing.SourceAppID.String, Valid: listing.SourceAppID.Valid},
		Status:             string(listing.Status),
		Items:              items,
		CommandCount:       int32(listing.CommandCount),
		EventListenerCount: int32(listing.EventListenerCount),
		MessageCount:       int32(listing.MessageCount),
		BlockTypes:         listing.BlockTypes,
		CreatedAt:          pgtype.Timestamp{Time: listing.CreatedAt.UTC(), Valid: true},
		UpdatedAt:          pgtype.Timestamp{Time: listing.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToMarketplaceListing(row)
}

func (c *Client) UpdateMarketplaceListing(ctx context.Context, listing *model.MarketplaceListing) (*model.MarketplaceListing, error) {
	items, err := json.Marshal(listing.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal items: %w", err)
	}

	row, err := c.Q.UpdateMarketplaceListing(ctx, pgmodel.UpdateMarketplaceListingParams{
		ID:                 listing.ID,
		Name:               listing.Name,
		Description:        listing.Description,
		SourceAppID:        pgtype.Text{String: listing.SourceAppID.String, Valid: listing.SourceAppID.Valid},
		Items:              items,
		CommandCount:       int32(listing.CommandCount),
		EventListenerCount: int32(listing.EventListenerCount),
		MessageCount:       int32(listing.MessageCount),
		BlockTypes:         listing.BlockTypes,
		Status:             string(listing.Status),
		UpdatedAt:          pgtype.Timestamp{Time: listing.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToMarketplaceListing(row)
}

func (c *Client) ReviewMarketplaceListing(
	ctx context.Context,
	id string,
	status model.MarketplaceListingStatus,
	note string,
	reviewerUserID string,
	reviewedAt time.Time,
) (*model.MarketplaceListing, error) {
	row, err := c.Q.ReviewMarketplaceListing(ctx, pgmodel.ReviewMarketplaceListingParams{
		ID:               id,
		Status:           string(status),
		ReviewNote:       pgtype.Text{String: note, Valid: note != ""},
		ReviewedByUserID: pgtype.Text{String: reviewerUserID, Valid: reviewerUserID != ""},
		ReviewedAt:       pgtype.Timestamp{Time: reviewedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToMarketplaceListing(row)
}

func (c *Client) IncrementMarketplaceListingImportCount(ctx context.Context, id string) error {
	return c.Q.IncrementMarketplaceListingImportCount(ctx, id)
}

func (c *Client) DeleteMarketplaceListing(ctx context.Context, id string) error {
	return c.Q.DeleteMarketplaceListing(ctx, id)
}

func (c *Client) MarketplaceListing(ctx context.Context, id string) (*model.MarketplaceListingWithAuthor, error) {
	row, err := c.Q.MarketplaceListing(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToMarketplaceListingWithAuthor(row.MarketplaceListing, row.User)
}

func (c *Client) MarketplaceListings(ctx context.Context, filter store.MarketplaceListingsFilter) ([]*model.MarketplaceListingWithAuthor, error) {
	rows, err := c.Q.MarketplaceListings(ctx, pgmodel.MarketplaceListingsParams{
		Status:     string(filter.Status),
		Search:     pgtype.Text{String: filter.Search, Valid: filter.Search != ""},
		Kind:       filter.Kind,
		Sort:       filter.Sort,
		PageOffset: int32(filter.Offset),
		PageLimit:  int32(filter.Limit),
	})
	if err != nil {
		return nil, err
	}

	res := make([]*model.MarketplaceListingWithAuthor, len(rows))
	for i, row := range rows {
		res[i], err = rowToMarketplaceListingWithAuthor(row.MarketplaceListing, row.User)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (c *Client) MarketplaceListingsByAuthor(ctx context.Context, authorUserID string) ([]*model.MarketplaceListingWithAuthor, error) {
	rows, err := c.Q.MarketplaceListingsByAuthor(ctx, authorUserID)
	if err != nil {
		return nil, err
	}

	res := make([]*model.MarketplaceListingWithAuthor, len(rows))
	for i, row := range rows {
		res[i], err = rowToMarketplaceListingWithAuthor(row.MarketplaceListing, row.User)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (c *Client) CountMarketplaceListingsByAuthor(ctx context.Context, authorUserID string) (int, error) {
	count, err := c.Q.CountMarketplaceListingsByAuthor(ctx, authorUserID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (c *Client) CreateMarketplaceReport(ctx context.Context, report *model.MarketplaceReport) (*model.MarketplaceReport, error) {
	row, err := c.Q.CreateMarketplaceReport(ctx, pgmodel.CreateMarketplaceReportParams{
		ID:             report.ID,
		ListingID:      report.ListingID,
		ReporterUserID: report.ReporterUserID,
		Reason:         report.Reason,
		CreatedAt:      pgtype.Timestamp{Time: report.CreatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToMarketplaceReport(row), nil
}

func (c *Client) CountOpenMarketplaceReportsByListing(ctx context.Context, listingID string) (int, error) {
	count, err := c.Q.CountOpenMarketplaceReportsByListing(ctx, listingID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (c *Client) OpenMarketplaceReports(ctx context.Context) ([]*model.MarketplaceReportWithReporter, error) {
	rows, err := c.Q.OpenMarketplaceReports(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*model.MarketplaceReportWithReporter, len(rows))
	for i, row := range rows {
		res[i] = &model.MarketplaceReportWithReporter{
			MarketplaceReport: *rowToMarketplaceReport(row.MarketplaceReport),
			Reporter:          rowToUser(row.User),
		}
	}

	return res, nil
}

func (c *Client) ResolveMarketplaceReport(ctx context.Context, id string, resolverUserID string, resolvedAt time.Time) error {
	return c.Q.ResolveMarketplaceReport(ctx, pgmodel.ResolveMarketplaceReportParams{
		ID:               id,
		ResolvedAt:       pgtype.Timestamp{Time: resolvedAt.UTC(), Valid: true},
		ResolvedByUserID: pgtype.Text{String: resolverUserID, Valid: resolverUserID != ""},
	})
}

func (c *Client) ResolveMarketplaceReportsByListing(ctx context.Context, listingID string, resolverUserID string, resolvedAt time.Time) error {
	return c.Q.ResolveMarketplaceReportsByListing(ctx, pgmodel.ResolveMarketplaceReportsByListingParams{
		ListingID:        listingID,
		ResolvedAt:       pgtype.Timestamp{Time: resolvedAt.UTC(), Valid: true},
		ResolvedByUserID: pgtype.Text{String: resolverUserID, Valid: resolverUserID != ""},
	})
}

func (c *Client) MarketplaceModerators(ctx context.Context) ([]*model.MarketplaceModerator, error) {
	rows, err := c.Q.MarketplaceModerators(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*model.MarketplaceModerator, len(rows))
	for i, row := range rows {
		res[i] = rowToMarketplaceModerator(row)
	}

	return res, nil
}

func (c *Client) MarketplaceModerator(ctx context.Context, discordUserID string) (*model.MarketplaceModerator, error) {
	row, err := c.Q.MarketplaceModerator(ctx, discordUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToMarketplaceModerator(row), nil
}

func (c *Client) CreateMarketplaceModerator(ctx context.Context, moderator *model.MarketplaceModerator) (*model.MarketplaceModerator, error) {
	row, err := c.Q.CreateMarketplaceModerator(ctx, pgmodel.CreateMarketplaceModeratorParams{
		DiscordUserID: moderator.DiscordUserID,
		AddedByUserID: pgtype.Text{String: moderator.AddedByUserID.String, Valid: moderator.AddedByUserID.Valid},
		CreatedAt:     pgtype.Timestamp{Time: moderator.CreatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToMarketplaceModerator(row), nil
}

func (c *Client) DeleteMarketplaceModerator(ctx context.Context, discordUserID string) error {
	return c.Q.DeleteMarketplaceModerator(ctx, discordUserID)
}

func rowToMarketplaceListingWithAuthor(row pgmodel.MarketplaceListing, author pgmodel.User) (*model.MarketplaceListingWithAuthor, error) {
	listing, err := rowToMarketplaceListing(row)
	if err != nil {
		return nil, err
	}

	return &model.MarketplaceListingWithAuthor{
		MarketplaceListing: *listing,
		Author:             rowToUser(author),
	}, nil
}

func rowToMarketplaceListing(row pgmodel.MarketplaceListing) (*model.MarketplaceListing, error) {
	var items []model.MarketplaceListingItem
	if err := json.Unmarshal(row.Items, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal items: %w", err)
	}

	return &model.MarketplaceListing{
		ID:                 row.ID,
		Name:               row.Name,
		Description:        row.Description,
		AuthorUserID:       row.AuthorUserID,
		SourceAppID:        null.NewString(row.SourceAppID.String, row.SourceAppID.Valid),
		Status:             model.MarketplaceListingStatus(row.Status),
		Items:              items,
		CommandCount:       int(row.CommandCount),
		EventListenerCount: int(row.EventListenerCount),
		MessageCount:       int(row.MessageCount),
		BlockTypes:         row.BlockTypes,
		ImportCount:        int(row.ImportCount),
		ReviewNote:         null.NewString(row.ReviewNote.String, row.ReviewNote.Valid),
		ReviewedByUserID:   null.NewString(row.ReviewedByUserID.String, row.ReviewedByUserID.Valid),
		ReviewedAt:         null.NewTime(row.ReviewedAt.Time, row.ReviewedAt.Valid),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}, nil
}

func rowToMarketplaceReport(row pgmodel.MarketplaceReport) *model.MarketplaceReport {
	return &model.MarketplaceReport{
		ID:               row.ID,
		ListingID:        row.ListingID,
		ReporterUserID:   row.ReporterUserID,
		Reason:           row.Reason,
		CreatedAt:        row.CreatedAt.Time,
		ResolvedAt:       null.NewTime(row.ResolvedAt.Time, row.ResolvedAt.Valid),
		ResolvedByUserID: null.NewString(row.ResolvedByUserID.String, row.ResolvedByUserID.Valid),
	}
}

func rowToMarketplaceModerator(row pgmodel.MarketplaceModerator) *model.MarketplaceModerator {
	return &model.MarketplaceModerator{
		DiscordUserID: row.DiscordUserID,
		AddedByUserID: null.NewString(row.AddedByUserID.String, row.AddedByUserID.Valid),
		CreatedAt:     row.CreatedAt.Time,
	}
}
