package bundlereview

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// BundleReviewHandler lets the admins review users' bundles for the
// marketplace.
type BundleReviewHandler struct {
	bundleStore store.BundleStore
	userStore   store.UserStore
}

func NewBundleReviewHandler(bundleStore store.BundleStore, userStore store.UserStore) *BundleReviewHandler {
	return &BundleReviewHandler{bundleStore: bundleStore, userStore: userStore}
}

// HandlePendingBundleReviewList lists the reviews waiting for a decision,
// oldest first, without the content of their bundles.
func (h *BundleReviewHandler) HandlePendingBundleReviewList(c *handler.Context) (*wire.PendingBundleReviewListResponse, error) {
	reviews, err := h.bundleStore.PendingBundleReviews(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle reviews: %w", err)
	}

	res := make([]*wire.PendingBundleReview, len(reviews))
	for i, r := range reviews {
		res[i] = pendingBundleReviewToWire(r)
	}
	return &res, nil
}

// HandlePendingBundleReviewGet returns a review waiting for a decision with
// the content of the version it reviews.
func (h *BundleReviewHandler) HandlePendingBundleReviewGet(c *handler.Context) (*wire.PendingBundleReviewGetResponse, error) {
	r, err := h.bundleStore.PendingBundleReview(c.Context(), c.Param("reviewID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_bundle_review", "the review was decided or changed since, reload the page")
		}
		return nil, fmt.Errorf("failed to get bundle review: %w", err)
	}
	return pendingBundleReviewToWire(r), nil
}

// pendingBundleReviewToWire shows the bundle with the listing in review.
func pendingBundleReviewToWire(r *model.PendingBundleReview) *wire.PendingBundleReview {
	b := *r.Bundle
	b.Listing = &r.Review.Listing
	res := &wire.PendingBundleReview{
		BundleReview: *wire.BundleReviewToWire(r.Review),
		Bundle:       wire.BundleToWire(&b),
	}
	res.Bundle.Owner = wire.BundleOwnerToWire(r.Bundle.Owner)
	if v := r.Bundle.ListedVersion(); v > 0 {
		res.ListedVersion = null.IntFrom(int64(v))
	}
	if v := r.Bundle.ApprovedVersion; v > 0 {
		res.ApprovedVersion = null.IntFrom(int64(v))
	}
	return res
}

// HandleBundleReviewDecide approves a pending review, which lists its version
// in the marketplace, or rejects it with a reason for its owner.
func (h *BundleReviewHandler) HandleBundleReviewDecide(c *handler.Context, req wire.BundleReviewDecideRequest) (*wire.BundleReviewDecideResponse, error) {
	review := &model.BundleReview{
		ID:             c.Param("reviewID"),
		Status:         model.BundleReviewStatusRejected,
		Reason:         req.Reason,
		ReviewerUserID: null.StringFrom(c.Session.UserID),
		DecidedAt:      null.TimeFrom(time.Now().UTC()),
	}
	if req.Approve {
		review.Status = model.BundleReviewStatusApproved
		review.Reason = ""
	}

	decided, err := h.bundleStore.DecideBundleReview(c.Context(), review)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_bundle_review", "the review was decided or changed since, reload the page")
		}
		return nil, fmt.Errorf("failed to decide bundle review: %w", err)
	}
	return wire.BundleReviewToWire(decided), nil
}
