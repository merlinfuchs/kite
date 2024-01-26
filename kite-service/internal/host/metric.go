package host

import (
	"context"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

func (h HostEnvironment) TrackEventHandled(ctx context.Context, eventType string, success bool, totalDuration time.Duration, executionDuration time.Duration) {
	err := h.deploymentMetrics.CreateDeploymentMetricEntry(ctx, model.DeploymentMetricEntry{
		DeploymentID:       h.DeploymentID,
		Type:               model.DeploymentMetricEntryTypeEventHandled,
		EventType:          eventType,
		EventSuccess:       success,
		EventTotalTime:     totalDuration,
		EventExecutionTime: executionDuration,
		Timestamp:          time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Error creating deployment metric entry for event")
	}
}
