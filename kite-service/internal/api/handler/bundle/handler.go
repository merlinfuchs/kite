package bundle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type BundleHandler struct {
	bundleStore store.BundleStore
	userStore   store.UserStore
}

func NewBundleHandler(bundleStore store.BundleStore, userStore store.UserStore) *BundleHandler {
	return &BundleHandler{bundleStore: bundleStore, userStore: userStore}
}

// HandleBundleList lists the bundles of the marketplace, the official ones
// in their order and then the users' listed ones, newest first, with only
// what their cards show. It works without logging in, and marks the user's
// own ones.
func (h *BundleHandler) HandleBundleList(c *handler.Context) (*wire.BundleListResponse, error) {
	var userID string
	if c.Session != nil {
		userID = c.Session.UserID
	}

	list, err := h.loadList(c.Context())
	if err != nil {
		return nil, err
	}

	res := make([]*wire.Bundle, len(list))
	for i, b := range list {
		res[i] = wire.BundleToWire(b)
		if b.Owner != nil {
			res[i].Owner = wire.BundleOwnerToWire(b.Owner)
			res[i].Yours = b.OwnedBy(userID)
		}
	}
	return &res, nil
}

func (h *BundleHandler) loadList(ctx context.Context) ([]*model.Bundle, error) {
	official, err := h.bundleStore.OfficialBundles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundles: %w", err)
	}
	listed, err := h.bundleStore.ListedUserBundles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get listed bundles: %w", err)
	}

	// A bundle whose content can't be read is left out, not the whole list.
	list := make([]*model.Bundle, 0, len(official)+len(listed))
	for _, b := range append(official, listed...) {
		if b.Content, err = bundle.CardContent(b.Content); err != nil {
			slog.With("error", err, "bundle_id", b.ID).Error("Failed to get card content of bundle")
			continue
		}
		list = append(list, b)
	}
	return list, nil
}

// HandleBundleGet returns a bundle the user can see, which works without
// logging in for official and published ones. Users' bundles come with the
// name of who made them.
func (h *BundleHandler) HandleBundleGet(c *handler.Context) (*wire.BundleGetResponse, error) {
	var userID string
	if c.Session != nil {
		userID = c.Session.UserID
	}

	b, err := h.bundleStore.Bundle(c.Context(), c.Param("bundleID"))
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}
	if err == nil && b.RemovedAt.Valid && !b.OwnedBy(userID) {
		// Apps that added it can say why it's gone.
		return nil, handler.ErrNotFound("bundle_removed", "Kite took this template down for breaking the rules")
	}
	if err != nil || !b.VisibleTo(userID) {
		return nil, handler.ErrNotFound("unknown_bundle", "template not found")
	}

	res := wire.BundleToWire(b)
	res.Yours = b.OwnedBy(userID)
	if b.OwnerUserID.Valid {
		owner, err := h.userStore.User(c.Context(), b.OwnerUserID.String)
		if err != nil {
			return nil, fmt.Errorf("failed to get bundle owner: %w", err)
		}
		res.Owner = wire.BundleOwnerToWire(owner)
	}
	return res, nil
}

// HandleBundleStar stars a bundle of the marketplace for the user. Users
// can't star their own ones.
func (h *BundleHandler) HandleBundleStar(c *handler.Context) (*wire.BundleStarResponse, error) {
	b, err := h.visibleBundle(c)
	if err != nil {
		return nil, err
	}
	if !b.Listed {
		return nil, handler.ErrBadRequest("bundle_not_starrable", "only templates in the marketplace can be starred")
	}
	if b.OwnedBy(c.Session.UserID) {
		return nil, handler.ErrBadRequest("bundle_not_starrable", "you can't star your own templates")
	}

	count, err := h.bundleStore.StarBundle(c.Context(), b.ID, c.Session.UserID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to star bundle: %w", err)
	}
	return &wire.BundleStarResponse{StarCount: count}, nil
}

// HandleBundleUnstar takes the user's star from a bundle, also from one that
// left the marketplace since.
func (h *BundleHandler) HandleBundleUnstar(c *handler.Context) (*wire.BundleUnstarResponse, error) {
	b, err := h.visibleBundle(c)
	if err != nil {
		return nil, err
	}

	count, err := h.bundleStore.UnstarBundle(c.Context(), b.ID, c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to unstar bundle: %w", err)
	}
	return &wire.BundleUnstarResponse{StarCount: count}, nil
}

// HandleStarredBundleList returns the IDs of the bundles the user starred.
func (h *BundleHandler) HandleStarredBundleList(c *handler.Context) (*wire.StarredBundleListResponse, error) {
	ids, err := h.bundleStore.StarredBundleIDs(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get starred bundles: %w", err)
	}
	// An empty list, not null, for users who didn't star any.
	res := append(wire.StarredBundleListResponse{}, ids...)
	return &res, nil
}

// visibleBundle is the bundle of the request, if the logged in user can see
// it.
func (h *BundleHandler) visibleBundle(c *handler.Context) (*model.Bundle, error) {
	b, err := h.bundleStore.Bundle(c.Context(), c.Param("bundleID"))
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}
	if err != nil || !b.VisibleTo(c.Session.UserID) {
		return nil, handler.ErrNotFound("unknown_bundle", "template not found")
	}
	return b, nil
}

// HandleBundleReportCreate reports a user's bundle the user can see for
// breaking the rules. Reporting it again replaces their open report.
func (h *BundleHandler) HandleBundleReportCreate(c *handler.Context, req wire.BundleReportCreateRequest) (*wire.BundleReportCreateResponse, error) {
	b, err := h.visibleBundle(c)
	if err != nil {
		return nil, err
	}
	if !b.OwnerUserID.Valid || b.OwnedBy(c.Session.UserID) {
		return nil, handler.ErrBadRequest("bundle_not_reportable", "only templates of other users can be reported")
	}

	err = h.bundleStore.SubmitBundleReport(c.Context(), &model.BundleReport{
		ID:             util.UniqueID(),
		BundleID:       b.ID,
		ReporterUserID: c.Session.UserID,
		Category:       model.BundleReportCategory(req.Category),
		Details:        req.Details,
		CreatedAt:      time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to submit bundle report: %w", err)
	}
	return &wire.BundleReportCreateResponse{}, nil
}
