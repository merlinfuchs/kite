package host

import (
	"context"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/kite-sdk-go/log"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

func (h HostEnvironment) Log(ctx context.Context, level log.LogLevel, msg string) {
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
