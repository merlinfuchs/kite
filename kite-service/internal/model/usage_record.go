package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type UsageRecordType string

const (
	UsageRecordTypeFlowExecution UsageRecordType = "flow_execution"
)

type UsageRecord struct {
	ID              int64
	Type            UsageRecordType
	AppID           string
	CommandID       null.String
	EventListenerID null.String
	MessageID       null.String
	CreditsUsed     uint32
	CreatedAt       time.Time
}
