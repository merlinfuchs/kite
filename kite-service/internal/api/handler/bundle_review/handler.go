package bundlereview

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
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
	h.notifyOwner(c.Context(), decided.BundleID, func(discordID string, b *model.Bundle) {
		h.notifier.BundleReviewDecided(discordID, b, decided)
	})
	return wire.BundleReviewToWire(decided), nil
}

// notifyOwner calls send with the Discord ID of the owner of the bundle.
// A notification that can't be sent doesn't fail the request.
func (h *BundleReviewHandler) notifyOwner(ctx context.Context, bundleID string, send func(discordID string, b *model.Bundle)) {
	if h.notifier == nil {
		return
	}
	b, err := h.bundleStore.Bundle(ctx, bundleID)
	if err != nil {
		slog.Warn("Failed to get bundle to notify its owner", slog.String("bundle_id", bundleID), slog.String("error", err.Error()))
		return
	}
	owner, err := h.userStore.User(ctx, b.OwnerUserID.String)
	if err != nil {
		slog.Warn("Failed to get owner of bundle to notify", slog.String("bundle_id", bundleID), slog.String("error", err.Error()))
		return
	}
	send(owner.DiscordID, b)
}
