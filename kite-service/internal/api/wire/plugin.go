package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"gopkg.in/guregu/null.v4"
)

type Plugin struct {
	ID       string           `json:"id"`
	Metadata plugin.Metadata  `json:"metadata"`
	Config   plugin.Config    `json:"config"`
	Commands []plugin.Command `json:"commands"`
	Events   []plugin.Event   `json:"events"`
}

type PluginListResponse = []*Plugin

func PluginToWire(plugin plugin.Plugin) *Plugin {
	return &Plugin{
		ID:       plugin.ID(),
		Metadata: plugin.Metadata(),
		Config:   plugin.Config(),
		Commands: plugin.Commands(),
		Events:   plugin.Events(),
	}
}

type PluginInstance struct {
	ID                 string              `json:"id"`
	PluginID           string              `json:"plugin_id"`
	Enabled            bool                `json:"enabled"`
	AppID              string              `json:"app_id"`
	CreatorUserID      string              `json:"creator_user_id"`
	Config             plugin.ConfigValues `json:"config"`
	EnabledResourceIDs []string            `json:"enabled_resource_ids"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	LastDeployedAt     null.Time           `json:"last_deployed_at"`
}

type PluginInstanceGetResponse = PluginInstance

type PluginInstanceListResponse = []*PluginInstance

type PluginInstanceCreateRequest struct {
	PluginID           string              `json:"plugin_id"`
	Config             plugin.ConfigValues `json:"config"`
	EnabledResourceIDs []string            `json:"enabled_resource_ids"`
	Enabled            bool                `json:"enabled"`
}

func (req PluginInstanceCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.PluginID, validation.Required),
	)
}

type PluginInstanceCreateResponse = PluginInstance

type PluginInstanceUpdateRequest struct {
	Config             plugin.ConfigValues `json:"config"`
	EnabledResourceIDs []string            `json:"enabled_resource_ids"`
	Enabled            bool                `json:"enabled"`
}

func (req PluginInstanceUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req)
}

type PluginInstanceUpdateResponse = PluginInstance

type PluginInstanceUpdateEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

func (req PluginInstanceUpdateEnabledRequest) Validate() error {
	return nil
}

type PluginInstanceUpdateEnabledResponse = PluginInstance

type PluginInstanceDeleteResponse = Empty

func PluginInstanceToWire(pluginInstance *model.PluginInstance) *PluginInstance {
	if pluginInstance == nil {
		return nil
	}

	return &PluginInstance{
		ID:                 pluginInstance.ID,
		PluginID:           pluginInstance.PluginID,
		Enabled:            pluginInstance.Enabled,
		AppID:              pluginInstance.AppID,
		CreatorUserID:      pluginInstance.CreatorUserID,
		Config:             pluginInstance.Config,
		EnabledResourceIDs: pluginInstance.EnabledResourceIDs,
		CreatedAt:          pluginInstance.CreatedAt,
		UpdatedAt:          pluginInstance.UpdatedAt,
		LastDeployedAt:     pluginInstance.LastDeployedAt,
	}
}
