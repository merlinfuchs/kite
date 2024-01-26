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
		EventID:            d.EventID,
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
