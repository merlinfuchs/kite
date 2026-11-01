package bundlereview

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/core/notify"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// BundleReviewHandler lets the admins review users' bundles for the
// marketplace.
type BundleReviewHandler struct {
	bundleStore store.BundleStore
	userStore   store.UserStore
	notifier    *notify.Notifier
}

func NewBundleReviewHandler(bundleStore store.BundleStore, userStore store.UserStore, notifier *notify.Notifier) *BundleReviewHandler {
	return &BundleReviewHandler{bundleStore: bundleStore, userStore: userStore, notifier: notifier}
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

// HandleAutoRejectedBundleReviewList lists the reviews the AI pre-screen
// rejected, newest first, so the admins can reopen the ones it got wrong.
func (h *BundleReviewHandler) HandleAutoRejectedBundleReviewList(c *handler.Context) (*wire.AutoRejectedBundleReviewListResponse, error) {
	reviews, err := h.bundleStore.AutoRejectedBundleReviews(c.Context(), 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get auto-rejected bundle reviews: %w", err)
	}

	res := make([]*wire.PendingBundleReview, len(reviews))
	for i, r := range reviews {
		res[i] = pendingBundleReviewToWire(r)
	}
	return &res, nil
}

// HandleBundleReviewPrescreenRun has the AI pre-screen screen a pending
// review again, which it does within a minute.
func (h *BundleReviewHandler) HandleBundleReviewPrescreenRun(c *handler.Context) (*wire.BundleReviewPrescreenRunResponse, error) {
	if err := h.bundleStore.ClearBundleReviewPrescreen(c.Context(), c.Param("reviewID")); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_bundle_review", "the review was decided or changed since, reload the page")
		}
		return nil, fmt.Errorf("failed to clear bundle review pre-screen: %w", err)
	}
	return &wire.BundleReviewPrescreenRunResponse{}, nil
}

// HandleBundleReviewReopen puts a review the AI pre-screen rejected back in
// the queue.
func (h *BundleReviewHandler) HandleBundleReviewReopen(c *handler.Context) (*wire.BundleReviewReopenResponse, error) {
	if err := h.bundleStore.ReopenAutoRejectedBundleReview(c.Context(), c.Param("reviewID")); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_bundle_review", "the template has a newer version or review since, or was taken down")
		}
		return nil, fmt.Errorf("failed to reopen bundle review: %w", err)
	}
	return &wire.BundleReviewReopenResponse{}, nil
}

// pendingBundleReviewToWire shows the bundle with the listing in review, as
// it would be listed.
func pendingBundleReviewToWire(r *model.PendingBundleReview) *wire.PendingBundleReview {
	b := *r.Bundle
	b.ShowListing(&r.Review.Listing)
	b.AIUse = bundle.AIUseDisclosure(b.Content, r.Review.Listing.AIUse)
	b.Listed = true
	res := &wire.PendingBundleReview{
		BundleReview: *wire.BundleReviewToWire(r.Review),
		Bundle:       wire.BundleToWire(&b),
		OwnerUserID:  r.Bundle.OwnerUserID.String,
		Prescreen:    wire.BundleReviewPrescreenToWire(r.Review.Prescreen),
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
	if b := h.bundleToNotify(c.Context(), decided.BundleID); b != nil {
		h.notifier.BundleReviewDecided(b, decided)
	}
	return wire.BundleReviewToWire(decided), nil
}

// bundleToNotify returns the bundle with its owner to notify them, or nil if
// there's no one to notify. A notification that can't be sent doesn't fail
// the request.
func (h *BundleReviewHandler) bundleToNotify(ctx context.Context, bundleID string) *model.Bundle {
	if h.notifier == nil {
		return nil
	}
	bundles, err := h.bundleStore.BundleSummaries(ctx, []string{bundleID})
	if err != nil {
		slog.Warn("Failed to get bundle to notify its owner", slog.String("bundle_id", bundleID), slog.String("error", err.Error()))
		return nil
	}
	if len(bundles) == 0 || bundles[0].Owner == nil {
		return nil
	}
	return bundles[0]
}
