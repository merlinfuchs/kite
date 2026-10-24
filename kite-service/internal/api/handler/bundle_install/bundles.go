package bundleinstall

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

const (
	// maxBundleItems is the most commands, event listeners, messages and
	// variables a bundle can have together.
	maxBundleItems = 50
	// maxBundleSize is the most bytes the content of a version can have.
	maxBundleSize = 1 << 20
)

// HandleBundleCreate makes a bundle of the user from items of the app. It's
// the user's, so it stays when the app is deleted, but new versions can only
// be published from the app.
func (h *BundleInstallHandler) HandleBundleCreate(c *handler.Context, req wire.BundleCreateRequest) (*wire.BundleCreateResponse, error) {
	now := time.Now().UTC()
	b := &model.Bundle{
		ID:          util.UniqueID(),
		Name:        req.Name,
		Description: req.Description,
		OwnerUserID: null.StringFrom(c.Session.UserID),
		SourceAppID: null.StringFrom(c.App.ID),
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.setBundleContent(c, b, req.BundleExportRequest); err != nil {
		return nil, err
	}

	if err := h.bundleStore.CreateBundle(c.Context(), b, h.maxBundlesPerUser); err != nil {
		if errors.Is(err, store.ErrLimitReached) {
			return nil, handler.ErrBadRequest("too_many_bundles", fmt.Sprintf("you can have at most %d bundles", h.maxBundlesPerUser))
		}
		return nil, fmt.Errorf("failed to create bundle: %w", err)
	}
	return wire.UserBundleToWire(b), nil
}

// HandleBundleVersionCreate publishes items of the app as the new version of
// the user's bundle. Apps that installed it are offered the update.
func (h *BundleInstallHandler) HandleBundleVersionCreate(c *handler.Context, req wire.BundleVersionCreateRequest) (*wire.BundleVersionCreateResponse, error) {
	b, err := h.ownBundle(c)
	if err != nil {
		return nil, err
	}
	if b.SourceAppID.String != c.App.ID {
		return nil, handler.ErrBadRequest("wrong_app", "publish new versions from the app the bundle was made from")
	}

	fromVersion := b.Version
	b.Name = req.Name
	b.Description = req.Description
	b.Version++
	b.UpdatedAt = time.Now().UTC()
	if err := h.setBundleContent(c, b, req.BundleExportRequest); err != nil {
		return nil, err
	}

	if err := h.bundleStore.PublishBundleVersion(c.Context(), b, fromVersion); err != nil {
		if errors.Is(err, store.ErrOutdated) {
			return nil, handler.ErrBadRequest("bundle_changed", "the bundle changed, reload the page and try again")
		}
		return nil, fmt.Errorf("failed to publish bundle version: %w", err)
	}
	return wire.UserBundleToWire(b), nil
}

// setBundleContent exports the items of the app as the content of the
// version of the bundle. The items keep their keys between versions, they're
// their IDs in the app.
func (h *BundleInstallHandler) setBundleContent(c *handler.Context, b *model.Bundle, items wire.BundleExportRequest) error {
	content, err := h.export(c, items, b.Name)
	if err != nil {
		return err
	}
	content.ID = b.ID
	content.Version = b.Version
	content.Description = b.Description

	count := len(content.Commands) + len(content.EventListeners) + len(content.Messages) + len(content.Variables)
	if count > maxBundleItems {
		return handler.ErrBadRequest("too_many_items", fmt.Sprintf("a bundle can have at most %d items, with the message templates and variables they use it would have %d", maxBundleItems, count))
	}

	raw, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to encode bundle: %w", err)
	}
	if len(raw) > maxBundleSize {
		return handler.ErrBadRequest("bundle_too_large", fmt.Sprintf("the bundle is %d KB, it can be at most %d KB", len(raw)/1024, maxBundleSize/1024))
	}
	b.Content = raw
	b.ItemCount = count
	return nil
}

// HandleUserBundleList lists the bundles of the user, from all of their apps.
func (h *BundleInstallHandler) HandleUserBundleList(c *handler.Context) (*wire.UserBundleListResponse, error) {
	bundles, err := h.bundleStore.BundlesByOwner(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundles: %w", err)
	}

	res := make([]*wire.UserBundle, len(bundles))
	for i, b := range bundles {
		res[i] = wire.UserBundleToWire(b)
	}

	return &res, nil
}

// HandleBundleDelete deletes a bundle of the user. Apps that installed it
// keep what it added, but don't get updates anymore.
func (h *BundleInstallHandler) HandleBundleDelete(c *handler.Context) (*wire.BundleDeleteResponse, error) {
	b, err := h.ownBundle(c)
	if err != nil {
		return nil, err
	}

	if err := h.bundleStore.DeleteBundle(c.Context(), b.ID); err != nil {
		return nil, fmt.Errorf("failed to delete bundle: %w", err)
	}
	return &wire.BundleDeleteResponse{}, nil
}

// ownBundle returns the bundle of the route if the user made it. Others don't
// learn it exists.
func (h *BundleInstallHandler) ownBundle(c *handler.Context) (*model.Bundle, error) {
	b, err := h.bundleStore.Bundle(c.Context(), c.Param("bundleID"))
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}
	if err != nil || !b.OwnedBy(c.Session.UserID) {
		return nil, handler.ErrNotFound("unknown_bundle", "bundle not found")
	}
	return b, nil
}
