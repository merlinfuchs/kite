package bundle

import (
	"errors"
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type BundleHandler struct {
	bundleStore store.BundleStore
	userStore   store.UserStore
}

func NewBundleHandler(bundleStore store.BundleStore, userStore store.UserStore) *BundleHandler {
	return &BundleHandler{bundleStore: bundleStore, userStore: userStore}
}

// HandleBundleList lists the bundles of the marketplace, the official ones
// in their order and then the users' listed ones, newest first. It works
// without logging in, and marks the user's own ones.
func (h *BundleHandler) HandleBundleList(c *handler.Context) (*wire.BundleListResponse, error) {
	var userID string
	if c.Session != nil {
		userID = c.Session.UserID
	}

	official, err := h.bundleStore.OfficialBundles(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get bundles: %w", err)
	}
	listed, err := h.bundleStore.ListedUserBundles(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get listed bundles: %w", err)
	}

	res := make([]*wire.Bundle, 0, len(official)+len(listed))
	for _, b := range official {
		res = append(res, wire.BundleToWire(b))
	}
	for _, b := range listed {
		w := wire.BundleToWire(b)
		w.Owner = wire.BundleOwnerToWire(b.Owner)
		w.Yours = b.OwnedBy(userID)
		res = append(res, w)
	}
	return &res, nil
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
	if err != nil || !b.VisibleTo(userID) {
		return nil, handler.ErrNotFound("unknown_bundle", "bundle not found")
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
