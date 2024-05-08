package deployment

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

func (h *DeploymentHandler) HandleDeploymentLogEntryList(c *fiber.Ctx) error {
	entires, err := h.deploymentLogs.GetDeploymentLogEntries(c.Context(), c.Params("deploymentID"), c.Params("appID"))
	if err != nil {
		return err
	}

	res := make([]wire.DeploymentLogEntry, len(entires))
	for i, entry := range entires {
		res[i] = wire.DeploymentLogEntryToWire(&entry)
	}

	return c.JSON(wire.DeploymentLogEntryListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *DeploymentHandler) HandleDeploymentLogSummaryGet(c *fiber.Ctx) error {
	cutoff := time.Now().UTC().Add(-time.Hour * 24)

	summary, err := h.deploymentLogs.GetDeploymentLogSummary(c.Context(), c.Params("deploymentID"), c.Params("appID"), cutoff)
	if err != nil {
		return err
	}

	return c.JSON(wire.DeploymentLogSummaryGetResponse{
		Success: true,
		Data: wire.DeploymentLogSummary{
			DeploymentID: summary.DeploymentID,
			TotalCount:   summary.TotalCount,
			ErrorCount:   summary.ErrorCount,
			WarnCount:    summary.WarnCount,
			InfoCount:    summary.InfoCount,
			DebugCount:   summary.DebugCount,
		},
	})
}
