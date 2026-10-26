package bundle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// listCacheDuration is how long the marketplace is cached, which anyone can
// load. Reviews and takedowns show in it after that.
const listCacheDuration = time.Minute

// listRetryDuration is how long the error is returned when the first load
// failed.
const listRetryDuration = 5 * time.Second

type BundleHandler struct {
	bundleStore store.BundleStore
	userStore   store.UserStore

	listMu       sync.Mutex
	list         []*model.Bundle
	listCachedAt time.Time
	// listErr is why the last load failed while there was no list yet.
	listErr error
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

	list, err := h.cachedList(c.Context())
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

// cachedList is the marketplace with the content of the cards. The lock makes
// requests wait for one load when the cache expires. If it fails, the old
// list or the error is kept for another while, so the waiting requests don't
// each load it.
func (h *BundleHandler) cachedList(ctx context.Context) ([]*model.Bundle, error) {
	h.listMu.Lock()
	defer h.listMu.Unlock()
	if h.list != nil && time.Since(h.listCachedAt) < listCacheDuration {
		return h.list, nil
	}
	if h.list == nil && h.listErr != nil && time.Since(h.listCachedAt) < listRetryDuration {
		return nil, h.listErr
	}

	// The request that loads it may be canceled, the others still wait for it.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	list, err := h.loadList(ctx)
	h.listCachedAt = time.Now()
	if err != nil {
		if h.list == nil {
			h.listErr = err
			return nil, err
		}
		slog.With("error", err).Error("Failed to reload the marketplace, keeping the old one")
		return h.list, nil
	}
	h.list = list
	return list, nil
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

// HandleBundleReportCreate reports a user's bundle the user can see for
// breaking the rules. Reporting it again replaces their open report.
func (h *BundleHandler) HandleBundleReportCreate(c *handler.Context, req wire.BundleReportCreateRequest) (*wire.BundleReportCreateResponse, error) {
	b, err := h.bundleStore.Bundle(c.Context(), c.Param("bundleID"))
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}
	if err != nil || !b.VisibleTo(c.Session.UserID) {
		return nil, handler.ErrNotFound("unknown_bundle", "template not found")
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
