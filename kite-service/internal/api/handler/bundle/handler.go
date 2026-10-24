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
}

func NewBundleHandler(bundleStore store.BundleStore) *BundleHandler {
	return &BundleHandler{bundleStore: bundleStore}
}

// HandleBundleList lists the official bundles.
func (h *BundleHandler) HandleBundleList(c *handler.Context) (*wire.BundleListResponse, error) {
	bundles, err := h.bundleStore.OfficialBundles(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get bundles: %w", err)
	}

	res := make([]*wire.Bundle, len(bundles))
	for i, b := range bundles {
		res[i] = wire.BundleToWire(b)
	}

	return &res, nil
}

// HandleBundleGet returns a bundle the user can see, which works without
// logging in for official ones.
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
	return res, nil
}
