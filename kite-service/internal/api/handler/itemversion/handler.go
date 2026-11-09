package itemversion

import (
	"errors"
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// ItemVersionHandler lists the version history of commands, event listeners
// and messages. The stores record the versions when an item is changed.
type ItemVersionHandler struct {
	itemVersionStore store.ItemVersionStore
}

func NewItemVersionHandler(itemVersionStore store.ItemVersionStore) *ItemVersionHandler {
	return &ItemVersionHandler{
		itemVersionStore: itemVersionStore,
	}
}

func (h *ItemVersionHandler) HandleCommandVersionList(c *handler.Context) (*wire.ItemVersionListResponse, error) {
	return h.list(c, c.Command.ID)
}

func (h *ItemVersionHandler) HandleCommandVersionGet(c *handler.Context) (*wire.ItemVersionGetResponse, error) {
	return h.get(c, c.Command.ID)
}

func (h *ItemVersionHandler) HandleEventListenerVersionList(c *handler.Context) (*wire.ItemVersionListResponse, error) {
	return h.list(c, c.EventListener.ID)
}

func (h *ItemVersionHandler) HandleEventListenerVersionGet(c *handler.Context) (*wire.ItemVersionGetResponse, error) {
	return h.get(c, c.EventListener.ID)
}

func (h *ItemVersionHandler) HandleMessageVersionList(c *handler.Context) (*wire.ItemVersionListResponse, error) {
	return h.list(c, c.Message.ID)
}

func (h *ItemVersionHandler) HandleMessageVersionGet(c *handler.Context) (*wire.ItemVersionGetResponse, error) {
	return h.get(c, c.Message.ID)
}

func (h *ItemVersionHandler) list(c *handler.Context, targetID string) (*wire.ItemVersionListResponse, error) {
	versions, err := h.itemVersionStore.ItemVersionsByTarget(c.Context(), c.App.ID, targetID, model.MaxItemVersions)
	if err != nil {
		return nil, fmt.Errorf("failed to get item versions: %w", err)
	}

	res := make([]*wire.ItemVersion, len(versions))
	for i, version := range versions {
		res[i] = wire.ItemVersionToWire(version)
	}
	return &res, nil
}

func (h *ItemVersionHandler) get(c *handler.Context, targetID string) (*wire.ItemVersionGetResponse, error) {
	version, err := h.itemVersionStore.ItemVersion(c.Context(), c.App.ID, c.Param("versionID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_item_version", "Version not found")
		}
		return nil, fmt.Errorf("failed to get item version: %w", err)
	}
	if version.TargetID != targetID {
		return nil, handler.ErrNotFound("unknown_item_version", "Version not found")
	}

	return &wire.ItemVersionGetResponse{
		ID:            version.ID,
		Source:        string(version.Source),
		CreatorUserID: version.CreatorUserID,
		Content:       version.Content,
		CreatedAt:     version.CreatedAt,
	}, nil
}
