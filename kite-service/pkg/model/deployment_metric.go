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

type DeploymentMetricsSummary struct {
	FirstEntryAt time.Time
	LastEntryAt  time.Time

	TotalEventCount         int
	SuccessEventCount       int
	AvgEventExecutionTime   time.Duration
	TotalEventExecutionTime time.Duration
	AvgEventTotalTime       time.Duration
	TotalEventTotalTime     time.Duration

	TotalCallCount     int
	SuccessCallCount   int
	AvgCallTotalTime   time.Duration
	TotalCallTotalTime time.Duration
}

type DeploymentEventMetricEntry struct {
	Timestamp            time.Time
	TotalCount           int
	SuccessCount         int
	AverageExecutionTime time.Duration
	AverageTotalTime     time.Duration
}

type DeploymentCallMetricEntry struct {
	Timestamp        time.Time
	TotalCount       int
	SuccessCount     int
	AverageTotalTime time.Duration
}
