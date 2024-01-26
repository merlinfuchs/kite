package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Deployment struct {
	ID                    string
	Key                   string
	Name                  string
	Description           string
	GuildID               string
	PluginVersionID       null.String
	WasmBytes             []byte
	ManifestDefaultConfig map[string]string
	ManifestEvents        []string
	ManifestCommands      []string
	Config                map[string]string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
