package model

import (
	"encoding/json"
	"time"

	"gopkg.in/guregu/null.v4"
)

type PluginInstance struct {
	AppID              string
	PluginID           string
	Enabled            bool
	Config             json.RawMessage
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CommandsDeployedAt null.Time
}
