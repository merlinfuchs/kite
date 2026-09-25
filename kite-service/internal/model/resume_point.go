package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type ResumePoint struct {
	ID                string
	Type              ResumePointType
	AppID             string
	CommandID         null.String
	EventListenerID   null.String
	MessageID         null.String
	MessageInstanceID null.Int
	FlowSourceID      null.String
	FlowNodeID        string
	FlowState         flow.FlowContextState
	CreatedAt         time.Time
	ExpiresAt         null.Time
	LastUsedAt        time.Time

	// ResumeAt is only set for timers.
	ResumeAt null.Time
	// InteractionToken is encrypted. It's only set for timers that resume a
	// flow triggered by an interaction, so the flow can still respond to it.
	InteractionToken null.String
}

type ResumePointType string

const (
	ResumePointTypeModal             ResumePointType = "modal"
	ResumePointTypeMessageComponents ResumePointType = "message_components"
	ResumePointTypeTimer             ResumePointType = "timer"
)
