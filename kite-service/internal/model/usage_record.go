package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type UsageRecordType string

const (
	UsageRecordTypeCommandFlowExecution       UsageRecordType = "command_flow_execution"
	UsageRecordTypeEventListenerFlowExecution UsageRecordType = "event_listener_flow_execution"
	UsageRecordTypeMessageFlowExecution       UsageRecordType = "message_flow_execution"
)

type UsageRecord struct {
	ID              int64
	Type            UsageRecordType
	AppID           string
	CommandID       null.String
	EventListenerID null.String
	MessageID       null.String
	CreditsUsed     int
	CreatedAt       time.Time
}

type UsageCreditsUsedByType struct {
	Type        UsageRecordType
	CreditsUsed int
}

type UsageCreditsUsedByDay struct {
	Date        time.Time
	CreditsUsed int
}
