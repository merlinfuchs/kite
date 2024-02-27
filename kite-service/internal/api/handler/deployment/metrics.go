package deployment

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

func (h *DeploymentHandler) HandleDeploymentEventMetricsList(c *fiber.Ctx) error {
	startAt, groupBy, err := decodeTimeframe(c)
	if err != nil {
		return err
	}

	metrics, err := h.deploymentMetrics.GetDeploymentEventMetrics(c.Context(), c.Params("deploymentID"), startAt, groupBy)
	if err != nil {
		return err
	}

	res := make([]wire.DeploymentEventMetricEntry, len(metrics))
	for i, metric := range metrics {
		res[i] = wire.DeploymentEventMetricEntryToWire(&metric)
	}

	return c.JSON(wire.DeploymentMetricEventsListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *DeploymentHandler) HandleDeploymentCallMetricsList(c *fiber.Ctx) error {
	startAt, groupBy, err := decodeTimeframe(c)
	if err != nil {
		return err
	}

	metrics, err := h.deploymentMetrics.GetDeploymentCallMetrics(c.Context(), c.Params("deploymentID"), startAt, groupBy)
	if err != nil {
		return err
	}

	res := make([]wire.DeploymentCallMetricEntry, len(metrics))
	for i, metric := range metrics {
		res[i] = wire.DeploymentCallMetricEntryToWire(&metric)
	}

	return c.JSON(wire.DeploymentMetricCallsListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *DeploymentHandler) HandleDeploymentsEventMetricsList(c *fiber.Ctx) error {
	startAt, groupBy, err := decodeTimeframe(c)
	if err != nil {
		return err
	}

	metrics, err := h.deploymentMetrics.GetDeploymentsEventMetrics(
		c.Context(),
		distype.Snowflake(c.Params("guildID")),
		startAt,
		groupBy,
	)
	if err != nil {
		return err
	}

	res := make([]wire.DeploymentEventMetricEntry, len(metrics))
	for i, metric := range metrics {
		res[i] = wire.DeploymentEventMetricEntryToWire(&metric)
	}

	return c.JSON(wire.DeploymentMetricEventsListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *DeploymentHandler) HandleDeploymentsCallMetricsList(c *fiber.Ctx) error {
	startAt, groupBy, err := decodeTimeframe(c)
	if err != nil {
		return err
	}

	metrics, err := h.deploymentMetrics.GetDeploymentsCallMetrics(
		c.Context(),
		distype.Snowflake(c.Params("guildID")),
		startAt,
		groupBy,
	)
	if err != nil {
		return err
	}

	res := make([]wire.DeploymentCallMetricEntry, len(metrics))
	for i, metric := range metrics {
		res[i] = wire.DeploymentCallMetricEntryToWire(&metric)
	}

	return c.JSON(wire.DeploymentMetricCallsListResponse{
		Success: true,
		Data:    res,
	})
}

func decodeTimeframe(c *fiber.Ctx) (time.Time, time.Duration, error) {
	var startAt time.Time
	var groupBy time.Duration

	timeFrame := c.Query("timeframe", "day")
	switch timeFrame {
	case "hour":
		startAt = time.Now().UTC().Add(-time.Hour)
		groupBy = time.Minute * 5
	case "day":
		startAt = time.Now().UTC().Add(-time.Hour * 24)
		groupBy = time.Hour
	case "week":
		startAt = time.Now().UTC().Add(-time.Hour * 24 * 7)
		groupBy = time.Hour * 12
	case "month":
		startAt = time.Now().UTC().Add(-time.Hour * 24 * 30)
		groupBy = time.Hour * 24
	default:
		return time.Time{}, 0, helpers.BadRequest("invalid_timeframe", "The timeframe query parameter must be one of: hour, day, week, month")
	}

	return startAt, groupBy, nil
}
