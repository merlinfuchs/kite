package asset

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type AssetHandler struct {
	assetStore   store.AssetStore
	maxAssetSize int64
}

func NewAssetHandler(assetStore store.AssetStore, maxAssetSize int) *AssetHandler {
	return &AssetHandler{
		assetStore:   assetStore,
		maxAssetSize: int64(maxAssetSize),
	}
}

func (h *AssetHandler) HandleAssetCreate(c *handler.Context) (*wire.AssetCreateResponse, error) {
	file, header, err := c.FormFile("file", h.maxAssetSize)
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_form", "failed to get file from form")
	}

	if h.maxAssetSize != 0 && header.Size > h.maxAssetSize {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("file size exceeds maximum allowed size (%d)", h.maxAssetSize))
	}

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, handler.ErrBadRequest("invalid_content_type", "only images are allowed")
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	contentHash := util.HashBytes(content)

	asset, err := h.assetStore.CreateAsset(c.Context(), &model.Asset{
		ID:            util.UniqueID(),
		AppID:         c.App.ID,
		CreatorUserID: c.Session.UserID,
		Name:          header.Filename,
		ContentType:   contentType,
		ContentSize:   int(len(content)),
		ContentHash:   contentHash,
		Content:       content,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return wire.AssetToWire(asset), nil
}

func (h *AssetHandler) HandleAssetGet(c *handler.Context) (*wire.AssetGetResponse, error) {
	asset, err := h.assetStore.Asset(c.Context(), c.Param("assetID"))
	if err != nil {
		if err == store.ErrNotFound {
			return nil, handler.ErrNotFound("asset_not_found", "asset not found")
		}
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	if asset.AppID != c.App.ID {
		return nil, handler.ErrForbidden("missing_access", "asset does not belong to this app")
	}

	return wire.AssetToWire(asset), nil
}

func (h *AssetHandler) HandleAssetDownload(c *handler.Context) error {
	if c.Session == nil && c.Header("Referer") != "" {
		// We don't want people to use Kite as a CDN for their assets.
		// When the asset is used on an external site, the Referer header will be set and the session will be nil.
		// We have to allow unauthenticated access outside of websites to make assets work inside Discord.
		return handler.ErrUnauthorized("unauthorized", "session required")
	}

	asset, err := h.assetStore.AssetWithContent(c.Context(), c.Param("assetID"))
	if err != nil {
		if err == store.ErrNotFound {
			return handler.ErrNotFound("asset_not_found", "asset not found")
		}
		return fmt.Errorf("failed to get asset: %w", err)
	}

	c.SetHeader("Content-Type", asset.ContentType)
	c.SetHeader("Content-Disposition", "inline")

	return c.Send(http.StatusOK, asset.Content)
}
