package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type Command struct {
	ID             string
	Name           string
	Description    string
	Enabled        bool
	AppID          string
	ModuleID       null.String
	CreatorUserID  string
	FlowSource     flow.FlowData
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastDeployedAt null.Time
}
