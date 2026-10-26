package wire

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

// bundleCategories are the categories of the marketplace, like the ones of
// the templates of kite-web.
var bundleCategories = []any{"moderation", "ai", "community", "support", "games", "other"}

// bundleIconPattern matches the names of the icons of kite-web, which shows
// unknown ones as a package.
var bundleIconPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// BundleListingSubmitRequest submits the newest version of a bundle to
// the marketplace, with how it's shown there once approved.
type BundleListingSubmitRequest = BundleListing

func (req BundleListing) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Icon, validation.Required, validation.Length(1, 50), validation.Match(bundleIconPattern)),
		validation.Field(&req.Category, validation.Required, validation.In(bundleCategories...)),
		validation.Field(&req.About, validation.Required, validation.Length(1, 5), validation.Each(validation.Required, validation.Length(1, 1500))),
		// Users' bundles have no questions and answers for now.
		validation.Field(&req.FAQ, validation.Empty),
		validation.Field(&req.Setup, validation.Length(0, 5), validation.Each(validation.Required, validation.Length(1, 500))),
	)
}

type BundleListingSubmitResponse = BundleReview

// BundleReview is a review of a version of a user's bundle for the
// marketplace.
type BundleReview struct {
	ID      string        `json:"id"`
	Version int           `json:"version"`
	Listing BundleListing `json:"listing"`
	// Status is pending, approved or rejected.
	Status string `json:"status"`
	// Reason is why it was rejected.
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	DecidedAt null.Time `json:"decided_at"`
}

// PendingBundleReview is a review waiting for a decision, for the admins.
type PendingBundleReview struct {
	BundleReview `tstype:",extends"`
	// Bundle is the reviewed version with the submitted listing, as it would
	// be shown once approved. Its content is null in the list of reviews.
	Bundle *Bundle `json:"bundle"`
	// ListedVersion is the version listed in the marketplace, null if the
	// bundle isn't listed.
	ListedVersion null.Int `json:"listed_version"`
	// ApprovedVersion is the version apps install now, null if none was
	// approved yet. Approving the review replaces it.
	ApprovedVersion null.Int `json:"approved_version"`
}

type PendingBundleReviewListResponse = []*PendingBundleReview

type PendingBundleReviewGetResponse = PendingBundleReview

// BundleReviewDecideRequest approves or rejects a pending review.
type BundleReviewDecideRequest struct {
	Approve bool `json:"approve"`
	// Reason is why it's rejected, which its owner sees.
	Reason string `json:"reason"`
}

func (req BundleReviewDecideRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Reason, validation.When(!req.Approve, validation.Required), validation.Length(0, 1000)),
	)
}

type BundleReviewDecideResponse = BundleReview

func BundleListingToWire(listing model.BundleListing) BundleListing {
	faq := make([]BundleFAQ, len(listing.FAQ))
	for i, f := range listing.FAQ {
		faq[i] = BundleFAQ{Question: f.Question, Answer: f.Answer}
	}
	return BundleListing{
		Icon:     listing.Icon,
		Category: listing.Category,
		About:    listing.About,
		FAQ:      faq,
		Setup:    listing.Setup,
	}
}

func BundleListingFromWire(listing BundleListing) model.BundleListing {
	faq := make([]model.BundleFAQ, len(listing.FAQ))
	for i, f := range listing.FAQ {
		faq[i] = model.BundleFAQ{Question: f.Question, Answer: f.Answer}
	}
	return model.BundleListing{
		Icon:     listing.Icon,
		Category: listing.Category,
		About:    listing.About,
		FAQ:      faq,
		Setup:    listing.Setup,
	}
}

func BundleReviewToWire(review *model.BundleReview) *BundleReview {
	if review == nil {
		return nil
	}
	return &BundleReview{
		ID:        review.ID,
		Version:   review.Version,
		Listing:   BundleListingToWire(review.Listing),
		Status:    string(review.Status),
		Reason:    review.Reason,
		CreatedAt: review.CreatedAt,
		DecidedAt: review.DecidedAt,
	}
}
