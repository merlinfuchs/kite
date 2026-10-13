package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// Function is a reusable flow that commands, event listeners, message
// templates and other functions can run with the Call function block.
type Function struct {
	ID            string
	Name          string
	Description   string
	AppID         string
	ModuleID      null.String
	CreatorUserID string
	FlowSource    flow.FlowData
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
