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
}

type ResumePointType string

const (
	ResumePointTypeModal ResumePointType = "modal"
)
