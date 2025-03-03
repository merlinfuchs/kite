package wire

import (
	"encoding/json"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type Plugin struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Author      string       `json:"author"`
	Version     string       `json:"version"`
	Config      PluginConfig `json:"config"`
}

type PluginConfig struct {
	Sections []PluginConfigSection `json:"sections"`
}

type PluginConfigSection struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Fields      []PluginConfigField `json:"fields"`
}

type PluginConfigField struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	ItemType    string `json:"item_type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PluginInstance struct {
	AppID     string          `json:"app_id"`
	PluginID  string          `json:"plugin_id"`
	Enabled   bool            `json:"enabled"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type PluginListResponse = []Plugin

type PluginGetResponse struct {
	Plugin
	Instance *PluginInstance `json:"instance"`
}

type PluginUpdateRequest struct {
	Enabled bool            `json:"enabled"`
	Config  json.RawMessage `json:"config"`
}

type PluginUpdateResponse struct {
	Plugin
	Instance *PluginInstance `json:"instance"`
}

func PluginToWire(plugin plugin.Plugin) *Plugin {
	metadata := plugin.Metadata()

	return &Plugin{
		ID:          plugin.ID(),
		Name:        metadata.Name,
		Description: metadata.Description,
		Icon:        metadata.Icon,
		Author:      metadata.Author,
		Version:     plugin.Version(),
	}
}

func PluginInstanceToWire(instance *model.PluginInstance) *PluginInstance {
	if instance == nil {
		return nil
	}

	return &PluginInstance{
		AppID:     instance.AppID,
		PluginID:  instance.PluginID,
		Enabled:   instance.Enabled,
		Config:    instance.Config,
		CreatedAt: instance.CreatedAt,
		UpdatedAt: instance.UpdatedAt,
	}
}
