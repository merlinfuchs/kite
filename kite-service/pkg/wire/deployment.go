package wire

import (
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"gopkg.in/guregu/null.v4"
)

type Deployment struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	Key                   string            `json:"key"`
	Description           string            `json:"description"`
	GuildID               string            `json:"guild_id"`
	PluginVersionID       null.String       `json:"plugin_version_id"`
	WasmSize              int               `json:"wasm_size"`
	ManifestDefaultConfig map[string]string `json:"manifest_default_config"`
	ManifestEvents        []string          `json:"manifest_events"`
	ManifestCommands      []string          `json:"manifest_commands"`
	Config                map[string]string `json:"config"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

type DeploymentListResponse APIResponse[[]Deployment]

type DeploymentGetResponse APIResponse[Deployment]

type DeploymentCreateRequest struct {
	Key                   string            `json:"key"`
	Name                  string            `json:"name"`
	Description           string            `json:"description"`
	GuildID               string            `json:"guild_id"`
	PluginVersionID       null.String       `json:"plugin_version_id"`
	WasmBytes             string            `json:"wasm_bytes"`
	ManifestDefaultConfig map[string]string `json:"manifest_default_config"`
	ManifestEvents        []string          `json:"manifest_events"`
	ManifestCommands      []string          `json:"manifest_commands"`
	Config                map[string]string `json:"config"`
}

type DeploymentCreateResponse APIResponse[Deployment]

func DeploymentToWire(d *model.Deployment) Deployment {
	return Deployment{
		ID:                    d.ID,
		Name:                  d.Name,
		Key:                   d.Key,
		Description:           d.Description,
		GuildID:               d.GuildID,
		PluginVersionID:       d.PluginVersionID,
		WasmSize:              len(d.WasmBytes),
		ManifestDefaultConfig: d.ManifestDefaultConfig,
		ManifestEvents:        d.ManifestEvents,
		ManifestCommands:      d.ManifestCommands,
		Config:                d.Config,
		CreatedAt:             d.CreatedAt,
		UpdatedAt:             d.UpdatedAt,
	}
}
