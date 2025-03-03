package model

import (
	"encoding/json"
	"time"
)

type PluginInstance struct {
	AppID     string
	PluginID  string
	Enabled   bool
	Config    json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}
