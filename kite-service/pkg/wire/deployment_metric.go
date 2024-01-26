package wire

import (
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type DeploymentMetricEntry struct {
	ID                 uint64            `json:"id"`
	DeploymentID       string            `json:"deployment_id"`
	Type               string            `json:"type"`
	Metadata           map[string]string `json:"metadata"`
	EventID            uint64            `json:"event_id"`
	EventType          string            `json:"event_type"`
	EventSuccess       bool              `json:"event_succes"`
	EventExecutionTime time.Duration     `json:"event_execution_time"`
	EventTotalTime     time.Duration     `json:"event_total_time"`
	CallType           string            `json:"call_type"`
	CallSuccess        bool              `json:"call_succes"`
	CallTotalTime      time.Duration     `json:"call_total_time"`
	Timestamp          time.Time         `json:"timestamp"`
}

func DeploymentMetricEntryToWire(d *model.DeploymentMetricEntry) DeploymentMetricEntry {
	return DeploymentMetricEntry{
		ID:                 d.ID,
		DeploymentID:       d.DeploymentID,
		Type:               string(d.Type),
		Metadata:           d.Metadata,
		EventType:          d.EventType,
		EventSuccess:       d.EventSuccess,
		EventExecutionTime: d.EventExecutionTime,
		EventTotalTime:     d.EventTotalTime,
		CallType:           d.CallType,
		CallSuccess:        d.CallSuccess,
		CallTotalTime:      d.CallTotalTime,
		Timestamp:          d.Timestamp,
	}
}

type DeploymentEventMetricEntry struct {
	Timestamp            time.Time `json:"timestamp"`
	TotalCount           int       `json:"total_count"`
	SuccessCount         int       `json:"success_count"`
	AverageExecutionTime int64     `json:"average_execution_time"`
	AverageTotalTime     int64     `json:"average_total_time"`
}

func DeploymentEventMetricEntryToWire(d *model.DeploymentEventMetricEntry) DeploymentEventMetricEntry {
	return DeploymentEventMetricEntry{
		Timestamp:            d.Timestamp,
		TotalCount:           d.TotalCount,
		SuccessCount:         d.SuccessCount,
		AverageExecutionTime: d.AverageExecutionTime.Microseconds(),
		AverageTotalTime:     d.AverageTotalTime.Microseconds(),
	}
}

type DeploymentMetricEventsListResponse APIResponse[[]DeploymentEventMetricEntry]

type DeploymentCallMetricEntry struct {
	Timestamp        time.Time `json:"timestamp"`
	TotalCount       int       `json:"total_count"`
	SuccessCount     int       `json:"success_count"`
	AverageTotalTime int64     `json:"average_total_time"`
}

func DeploymentCallMetricEntryToWire(d *model.DeploymentCallMetricEntry) DeploymentCallMetricEntry {
	return DeploymentCallMetricEntry{
		Timestamp:        d.Timestamp,
		TotalCount:       d.TotalCount,
		SuccessCount:     d.SuccessCount,
		AverageTotalTime: d.AverageTotalTime.Microseconds(),
	}
}

type DeploymentMetricCallsListResponse APIResponse[[]DeploymentCallMetricEntry]
