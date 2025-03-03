package plugin

import (
	"errors"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type PluginHandler struct {
	registry            *plugin.Registry
	pluginInstanceStore store.PluginInstanceStore
	pluginValueStore    store.PluginValueStore
}

func NewPluginHandler(
	registry *plugin.Registry,
	pluginInstanceStore store.PluginInstanceStore,
	pluginValueStore store.PluginValueStore,
) *PluginHandler {
	return &PluginHandler{
		registry:            registry,
		pluginInstanceStore: pluginInstanceStore,
		pluginValueStore:    pluginValueStore,
	}
}

func (h *PluginHandler) HandlePluginList(c *handler.Context) (*wire.PluginListResponse, error) {
	plugins := h.registry.Plugins()

	res := make(wire.PluginListResponse, len(plugins))
	for i, plugin := range plugins {
		res[i] = *wire.PluginToWire(plugin)
	}

	return &res, nil
}

func (h *PluginHandler) HandlePluginGet(c *handler.Context) (*wire.PluginGetResponse, error) {
	pluginID := c.Param("pluginID")

	plugin := h.registry.Plugin(pluginID)
	if plugin == nil {
		return nil, handler.ErrNotFound("unknown_plugin", "Unknown plugin")
	}

	instance, err := h.pluginInstanceStore.PluginInstance(c.Context(), c.App.ID, pluginID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	return &wire.PluginGetResponse{
		Plugin:   *wire.PluginToWire(plugin),
		Instance: wire.PluginInstanceToWire(instance),
	}, nil
}

func (h *PluginHandler) HandlePluginInstanceUpdate(c *handler.Context, req *wire.PluginUpdateRequest) (*wire.PluginUpdateResponse, error) {
	pluginID := c.Param("pluginID")
	appID := c.Param("appID")

	plugin := h.registry.Plugin(pluginID)
	if plugin == nil {
		return nil, handler.ErrNotFound("unknown_plugin", "Unknown plugin")
	}

	instance, err := h.pluginInstanceStore.UpsertPluginInstance(c.Context(), model.PluginInstance{
		AppID:     appID,
		PluginID:  pluginID,
		Enabled:   req.Enabled,
		Config:    req.Config,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	return &wire.PluginUpdateResponse{
		Plugin:   *wire.PluginToWire(plugin),
		Instance: wire.PluginInstanceToWire(instance),
	}, nil
}
