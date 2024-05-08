package model

import (
	"time"

	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
	"gopkg.in/guregu/null.v4"
)

type Deployment struct {
	ID              string
	Key             string
	Name            string
	Description     string
	AppID           string
	PluginVersionID null.String
	WasmBytes       []byte
	Manifest        manifest.Manifest
	Config          map[string]string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeployedAt      null.Time
}

type PartialDeployment struct {
	ID    string
	AppID string
}
