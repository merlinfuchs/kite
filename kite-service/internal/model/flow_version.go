package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// FlowVersion is an earlier save of the flow of a command or event listener.
type FlowVersion struct {
	ID              string
	AppID           string
	CommandID       null.String
	EventListenerID null.String
	// Empty in lists, which only load it for a single version.
	FlowSource    flow.FlowData
	AutoSaved     bool
	CreatorUserID null.String
	// Only set in lists.
	CreatorDisplayName null.String
	CreatedAt          time.Time
}
