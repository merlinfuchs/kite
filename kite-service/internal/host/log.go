package host

import (
	"context"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-types/logmodel"
)

func (h HostEnvironment) Log(ctx context.Context, level logmodel.LogLevel, msg string) {
	err := h.deploymentLogs.CreateDeploymentLogEntry(ctx, model.DeploymentLogEntry{
		DeploymentID: h.DeploymentID,
		Level:        level.Name(),
		Message:      msg,
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Error creating deployment log entry")
	}
}
