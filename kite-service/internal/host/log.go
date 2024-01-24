package host

import (
	"context"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/go-types/logmodel"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/internal/util"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

func (h HostEnvironment) Log(ctx context.Context, deploymentID string, level logmodel.LogLevel, msg string) {
	err := h.deployments.CreateDeploymentLogEntry(ctx, model.DeploymentLogEntry{
		ID:           util.UniqueID(),
		DeploymentID: deploymentID,
		Level:        level.Name(),
		Message:      msg,
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Error creating deployment log entry")
	}
}
