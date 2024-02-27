package wire

import (
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

// DeploymentEventMetricEntry represents a single deployment event metric entry
// All times are in milliseconds
type DeploymentEventMetricEntry struct {
	Timestamp            time.Time `json:"timestamp"`
	TotalCount           int       `json:"total_count"`
	SuccessCount         int       `json:"success_count"`
	AverageExecutionTime float32   `json:"average_execution_time"`
	AverageTotalTime     float32   `json:"average_total_time"`
}

func DeploymentEventMetricEntryToWire(d *model.DeploymentEventMetricEntry) DeploymentEventMetricEntry {
	return DeploymentEventMetricEntry{
		Timestamp:            d.Timestamp,
		TotalCount:           d.TotalCount,
		SuccessCount:         d.SuccessCount,
		AverageExecutionTime: float32(d.AverageExecutionTime.Microseconds()) / 1000,
		AverageTotalTime:     float32(d.AverageTotalTime.Microseconds()) / 1000,
	}
}

type DeploymentMetricEventsListResponse APIResponse[[]DeploymentEventMetricEntry]

// DeploymentCallMetricEntry represents a single deployment call metric entry
// All times are in milliseconds
type DeploymentCallMetricEntry struct {
	Timestamp        time.Time `json:"timestamp"`
	TotalCount       int       `json:"total_count"`
	SuccessCount     int       `json:"success_count"`
	AverageTotalTime float32   `json:"average_total_time"`
}

func DeploymentCallMetricEntryToWire(d *model.DeploymentCallMetricEntry) DeploymentCallMetricEntry {
	return DeploymentCallMetricEntry{
		Timestamp:        d.Timestamp,
		TotalCount:       d.TotalCount,
		SuccessCount:     d.SuccessCount,
		AverageTotalTime: float32(d.AverageTotalTime.Microseconds()) / 1000,
	}
}

type DeploymentMetricCallsListResponse APIResponse[[]DeploymentCallMetricEntry]
