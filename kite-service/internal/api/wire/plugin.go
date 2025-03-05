package wire

import (
	"encoding/json"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"gopkg.in/guregu/null.v4"
)

type Plugin struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Author      string       `json:"author"`
	Config      PluginConfig `json:"config"`
	Default     bool         `json:"default"`
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
	AppID              string          `json:"app_id"`
	PluginID           string          `json:"plugin_id"`
	Enabled            bool            `json:"enabled"`
	Config             json.RawMessage `json:"config"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	CommandsDeployedAt null.Time       `json:"commands_deployed_at"`
}

type PluginListResponse = []Plugin

type PluginInstanceGetResponse = PluginInstance
type PluginInstanceUpdateRequest struct {
	Enabled bool            `json:"enabled"`
	Config  json.RawMessage `json:"config"`
}

type PluginInstanceUpdateResponse = PluginInstance

func PluginToWire(plugin plugin.Plugin) *Plugin {
	metadata := plugin.Metadata()

	return &Plugin{
		ID:          plugin.ID(),
		Name:        metadata.Name,
		Description: metadata.Description,
		Icon:        metadata.Icon,
		Author:      metadata.Author,
		Default:     plugin.IsDefault(),
	}
}

func PluginInstanceToWire(instance *model.PluginInstance) *PluginInstance {
	if instance == nil {
		return nil
	}

	return &PluginInstance{
		AppID:              instance.AppID,
		PluginID:           instance.PluginID,
		Enabled:            instance.Enabled,
		Config:             instance.Config,
		CreatedAt:          instance.CreatedAt,
		UpdatedAt:          instance.UpdatedAt,
		CommandsDeployedAt: instance.CommandsDeployedAt,
	}
}
