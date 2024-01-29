package model

import (
	"time"

	"github.com/merlinfuchs/kite/go-types/manifest"
	"gopkg.in/guregu/null.v4"
)

type Deployment struct {
	ID              string
	Key             string
	Name            string
	Description     string
	GuildID         string
	PluginVersionID null.String
	WasmBytes       []byte
	Manifest        manifest.Manifest
	Config          map[string]string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
