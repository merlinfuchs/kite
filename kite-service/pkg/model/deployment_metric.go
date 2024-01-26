package model

import (
	"time"
)

type DeploymentMetricEntryType string

const (
	DeploymentMetricEntryTypeEventHandled DeploymentMetricEntryType = "EVENT_HANDLED"
	DeploymentMetricEntryTypeCallExecuted DeploymentMetricEntryType = "CALL_EXECUTED"
)

type DeploymentMetricEntry struct {
	ID                 uint64
	DeploymentID       string
	Type               DeploymentMetricEntryType
	Metadata           map[string]string
	EventType          string
	EventSuccess       bool
	EventExecutionTime time.Duration
	EventTotalTime     time.Duration
	CallType           string
	CallSuccess        bool
	CallTotalTime      time.Duration
	Timestamp          time.Time
}
