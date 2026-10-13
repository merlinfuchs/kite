package plugin

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/command"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type PluginHandler struct {
	pluginRegistry      *plugin.Registry
	pluginInstanceStore store.PluginInstanceStore
	commandManager      *command.CommandManager
}

func NewPluginHandler(
	pluginRegistry *plugin.Registry,
	pluginInstanceStore store.PluginInstanceStore,
	commandManager *command.CommandManager,
) *PluginHandler {
	return &PluginHandler{
		pluginRegistry:      pluginRegistry,
		pluginInstanceStore: pluginInstanceStore,
		commandManager:      commandManager,
	}
}

func (h *PluginHandler) HandlePluginList(c *handler.Context) (*wire.PluginListResponse, error) {
	plugins := h.pluginRegistry.Plugins()

	res := make([]*wire.Plugin, len(plugins))
	for i, plugin := range plugins {
		res[i] = wire.PluginToWire(plugin)
	}

	return &res, nil
}

func (h *PluginHandler) HandlePluginInstanceList(c *handler.Context) (*wire.PluginInstanceListResponse, error) {
	pluginInstances, err := h.pluginInstanceStore.PluginInstancesByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin instances: %w", err)
	}

	res := make([]*wire.PluginInstance, len(pluginInstances))
	for i, pluginInstance := range pluginInstances {
		res[i] = wire.PluginInstanceToWire(pluginInstance)
	}

	return &res, nil
}

func (h *PluginHandler) HandlePluginInstanceGet(c *handler.Context) (*wire.PluginInstanceGetResponse, error) {
	return wire.PluginInstanceToWire(c.PluginInstance), nil
}

func (h *PluginHandler) HandlePluginInstanceCreate(c *handler.Context, req wire.PluginInstanceCreateRequest) (*wire.PluginInstanceCreateResponse, error) {
	pluginInstance, err := h.pluginInstanceStore.CreatePluginInstance(c.Context(), &model.PluginInstance{
		ID:                 util.UniqueID(),
		PluginID:           req.PluginID,
		AppID:              c.App.ID,
		CreatorUserID:      c.Session.UserID,
		Config:             req.Config,
		EnabledResourceIDs: req.EnabledResourceIDs,
		Enabled:            req.Enabled,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin instance: %w", err)
	}

	h.markCommandsChanged(c)

	return wire.PluginInstanceToWire(pluginInstance), nil
}

func (h *PluginHandler) HandlePluginInstanceUpdate(c *handler.Context, req wire.PluginInstanceUpdateRequest) (*wire.PluginInstanceUpdateResponse, error) {
	pluginInstance, err := h.pluginInstanceStore.UpdatePluginInstance(c.Context(), &model.PluginInstance{
		ID:                 c.PluginInstance.ID,
		AppID:              c.App.ID,
		PluginID:           c.PluginInstance.PluginID,
		Config:             req.Config,
		EnabledResourceIDs: req.EnabledResourceIDs,
		Enabled:            req.Enabled,
		UpdatedAt:          time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_plugin_instance", "Plugin instance not found")
		}
		return nil, fmt.Errorf("failed to update plugin instance: %w", err)
	}

	h.markCommandsChanged(c)

	return wire.PluginInstanceToWire(pluginInstance), nil
}

func (h *PluginHandler) HandlePluginInstanceUpdateEnabled(c *handler.Context, req wire.PluginInstanceUpdateEnabledRequest) (*wire.PluginInstanceUpdateEnabledResponse, error) {
	pluginInstance, err := h.pluginInstanceStore.UpdatePluginInstance(c.Context(), &model.PluginInstance{
		ID:                 c.PluginInstance.ID,
		AppID:              c.App.ID,
		PluginID:           c.PluginInstance.PluginID,
		Config:             c.PluginInstance.Config,
		EnabledResourceIDs: c.PluginInstance.EnabledResourceIDs,
		Enabled:            req.Enabled,
		UpdatedAt:          time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_plugin_instance", "Plugin instance not found")
		}
		return nil, fmt.Errorf("failed to update plugin instance: %w", err)
	}

	return wire.PluginInstanceToWire(pluginInstance), nil
}

func (h *PluginHandler) HandlePluginInstanceDelete(c *handler.Context) (*wire.PluginInstanceDeleteResponse, error) {
	err := h.pluginInstanceStore.DeletePluginInstance(c.Context(), c.App.ID, c.PluginInstance.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_plugin_instance", "Plugin instance not found")
		}
		return nil, fmt.Errorf("failed to delete plugin instance: %w", err)
	}

	h.markCommandsChanged(c)

	return &wire.PluginInstanceDeleteResponse{}, nil
}

// markCommandsChanged queues the commands to be deployed, as plugins can add
// commands. The change is saved already, so a failure is only logged.
func (h *PluginHandler) markCommandsChanged(c *handler.Context) {
	if err := h.commandManager.MarkCommandsChanged(c.Context(), c.App.ID, false); err != nil {
		slog.Error(
			"Failed to mark commands changed",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
	}
}
