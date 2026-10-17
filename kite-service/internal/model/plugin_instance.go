package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type PluginInstance struct {
	ID                 string
	PluginID           string
	Enabled            bool
	AppID              string
	CreatorUserID      string
	Config             plugin.ConfigValues
	EnabledResourceIDs []string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
