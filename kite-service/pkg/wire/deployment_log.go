package wire

import (
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type DeploymentLogEntry struct {
	ID           uint64    `json:"id"`
	DeploymentID string    `json:"deployment_id"`
	Level        string    `json:"level"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}

type DeploymentLogEntryListResponse APIResponse[[]DeploymentLogEntry]

type DeploymentLogSummary struct {
	DeploymentID string `json:"deployment_id"`
	TotalCount   int    `json:"total_count"`
	ErrorCount   int    `json:"error_count"`
	WarnCount    int    `json:"warn_count"`
	InfoCount    int    `json:"info_count"`
	DebugCount   int    `json:"debug_count"`
}

type DeploymentLogSummaryGetResponse APIResponse[DeploymentLogSummary]

func DeploymentLogEntryToWire(d *model.DeploymentLogEntry) DeploymentLogEntry {
	return DeploymentLogEntry{
		ID:           d.ID,
		DeploymentID: d.DeploymentID,
		Level:        d.Level,
		Message:      d.Message,
		CreatedAt:    d.CreatedAt,
	}
}
