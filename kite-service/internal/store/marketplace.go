package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type MarketplaceListingsFilter struct {
	Status model.MarketplaceListingStatus
	Search string
	// Kind is "", "command", "event_listener" or "module".
	Kind string
	// Sort is "popular" or "recent".
	Sort   string
	Offset int
	Limit  int
}

type MarketplaceStore interface {
	CreateMarketplaceListing(ctx context.Context, listing *model.MarketplaceListing) (*model.MarketplaceListing, error)
	// UpdateMarketplaceListing replaces the content and clears the last review.
	UpdateMarketplaceListing(ctx context.Context, listing *model.MarketplaceListing) (*model.MarketplaceListing, error)
	ReviewMarketplaceListing(ctx context.Context, id string, status model.MarketplaceListingStatus, note string, reviewerUserID string, reviewedAt time.Time) (*model.MarketplaceListing, error)
	IncrementMarketplaceListingImportCount(ctx context.Context, id string) error
	DeleteMarketplaceListing(ctx context.Context, id string) error
	MarketplaceListing(ctx context.Context, id string) (*model.MarketplaceListingWithAuthor, error)
	MarketplaceListings(ctx context.Context, filter MarketplaceListingsFilter) ([]*model.MarketplaceListingWithAuthor, error)
	MarketplaceListingsByAuthor(ctx context.Context, authorUserID string) ([]*model.MarketplaceListingWithAuthor, error)
	CountMarketplaceListingsByAuthor(ctx context.Context, authorUserID string) (int, error)

	// CreateMarketplaceReport updates the reason of the user's open report
	// on the listing instead if there already is one.
	CreateMarketplaceReport(ctx context.Context, report *model.MarketplaceReport) (*model.MarketplaceReport, error)
	CountOpenMarketplaceReportsByListing(ctx context.Context, listingID string) (int, error)
	OpenMarketplaceReports(ctx context.Context) ([]*model.MarketplaceReportWithReporter, error)
	ResolveMarketplaceReport(ctx context.Context, id string, resolverUserID string, resolvedAt time.Time) error
	ResolveMarketplaceReportsByListing(ctx context.Context, listingID string, resolverUserID string, resolvedAt time.Time) error

	MarketplaceModerators(ctx context.Context) ([]*model.MarketplaceModerator, error)
	MarketplaceModerator(ctx context.Context, discordUserID string) (*model.MarketplaceModerator, error)
	CreateMarketplaceModerator(ctx context.Context, moderator *model.MarketplaceModerator) (*model.MarketplaceModerator, error)
	DeleteMarketplaceModerator(ctx context.Context, discordUserID string) error
}
