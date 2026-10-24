package bundle

import (
	"errors"
	"fmt"
	"slices"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
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
	bundles, versions, err := h.bundleStore.OfficialBundlesCurrentVersion(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get bundles: %w", err)
	}

	res := make([]*wire.Bundle, len(bundles))
	for i, b := range bundles {
		res[i] = wire.BundleToWire(b, versions[i])
	}
	slices.SortStableFunc(res, func(a, b *wire.Bundle) int {
		return bundle.OfficialPosition(a.ID) - bundle.OfficialPosition(b.ID)
	})

	return &res, nil
}

// HandleBundleGet returns a published bundle. Others are only shared by share
// code.
func (h *BundleHandler) HandleBundleGet(c *handler.Context) (*wire.BundleGetResponse, error) {
	b, version, err := h.bundleStore.ListedBundleCurrentVersion(c.Context(), c.Param("bundleID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_bundle", "bundle not found")
		}
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}

	return wire.BundleToWire(b, version), nil
}
