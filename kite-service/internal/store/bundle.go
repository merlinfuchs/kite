package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type BundleStore interface {
	// Bundle returns the bundle with the content of the version apps install,
	// its listing, if it has one, and the notes of its versions.
	Bundle(ctx context.Context, id string) (*model.Bundle, error)
	// OfficialBundles returns the listed official bundles in their order.
	OfficialBundles(ctx context.Context) ([]*model.Bundle, error)
	// BundleSummaries returns the bundles with the IDs that exist, without
	// their content and with only the icon of their listing, and their owner.
	BundleSummaries(ctx context.Context, ids []string) ([]*model.Bundle, error)
	// ListedUserBundles returns the users' bundles listed in the marketplace,
	// newest first, as their listed version and with their owner.
	ListedUserBundles(ctx context.Context) ([]*model.Bundle, error)
	// UpsertOfficialBundle stores the bundle with its listing and makes its
	// content the version apps install. The content of a stored version is
	// replaced, and so are the notes of the versions in VersionNotes.
	UpsertOfficialBundle(ctx context.Context, bundle *model.Bundle) error
	// BundlesByOwner returns the user's bundles, newest first, without their
	// content but with how many items it has.
	BundlesByOwner(ctx context.Context, userID string) ([]*model.Bundle, error)
	// CreateBundle creates a user's bundle with its first version. It returns
	// ErrLimitReached if the user has maxPerOwner bundles already.
	CreateBundle(ctx context.Context, bundle *model.Bundle, maxPerOwner int) error
	// PublishBundleVersion stores the content as the bundle's new version and
	// makes it the version apps install, keeping a few versions before it. A
	// bundle a review approved keeps its approved version until a review
	// approves the new one, which is submitted while it's listed. It returns
	// ErrOutdated if the bundle isn't at fromVersion anymore.
	PublishBundleVersion(ctx context.Context, bundle *model.Bundle, fromVersion int) error
	// DeleteBundle deletes the bundle with its versions and listing. Apps
	// that installed it keep what it added. It returns ErrNotFound if Kite
	// took it down, which keeps it.
	DeleteBundle(ctx context.Context, id string) error
	// UnlistBundle takes the bundle out of the marketplace and ends its
	// pending review. Apps stay on its approved version, and listing it again
	// needs a new review.
	UnlistBundle(ctx context.Context, id string) error
	// SubmitBundleReview submits the newest version of the bundle to the
	// marketplace, replacing its pending review, and sets the version of the
	// review. It returns ErrNotFound if the bundle was taken down.
	SubmitBundleReview(ctx context.Context, review *model.BundleReview) error
	// PendingBundleReviews returns the reviews waiting for a decision, oldest
	// first, with their bundle without its content and the owner.
	PendingBundleReviews(ctx context.Context) ([]*model.PendingBundleReview, error)
	// PendingBundleReview returns the pending review with its ID, with the
	// content of the version it reviews and the owner. It returns ErrNotFound
	// if it isn't pending.
	PendingBundleReview(ctx context.Context, id string) (*model.PendingBundleReview, error)
	// DecideBundleReview stores the status, reason, reviewer and time of the
	// decision on the pending review with its ID, and lists its version if
	// it's approved. It returns the decided review, or ErrNotFound if it isn't
	// pending. A review gets a new ID when its version or listing changes, so
	// a decision is always on what the admin saw.
	DecideBundleReview(ctx context.Context, decision *model.BundleReview) (*model.BundleReview, error)
	// SubmitBundleReport reports a bundle, replacing the open report of the
	// same user.
	SubmitBundleReport(ctx context.Context, report *model.BundleReport) error
	// OpenBundleReports returns the reports waiting for a decision, oldest
	// first, with their bundle without its content.
	OpenBundleReports(ctx context.Context) ([]*model.OpenBundleReport, error)
	// DismissBundleReport resolves the open report without taking its bundle
	// down. It returns ErrNotFound if the report isn't open.
	DismissBundleReport(ctx context.Context, id string, resolverUserID string, at time.Time) error
	// RemoveBundle takes a user's bundle down for breaking the rules, which
	// hides it from others, takes it out of the marketplace and resolves its
	// open reports. It returns ErrNotFound if it's official or taken down already.
	RemoveBundle(ctx context.Context, id string, reason string, resolverUserID string, at time.Time) error
}
