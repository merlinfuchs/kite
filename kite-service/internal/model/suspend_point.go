package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type SuspendPoint struct {
	ID              string
	Type            SuspendPointType
	AppID           string
	CommandID       null.String
	EventListenerID null.String
	MessageID       null.String
	FlowSourceID    null.String
	FlowNodeID      string
	FlowState       flow.FlowContextState
	CreatedAt       time.Time
	ExpiresAt       null.Time
}

type SuspendPointType string

const (
	SuspendPointTypeModal SuspendPointType = "modal"
)
