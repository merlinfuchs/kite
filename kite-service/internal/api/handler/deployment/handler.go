package deployment

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/config"
	"github.com/merlinfuchs/kite/kite-service/internal/util"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/module"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type DeploymentHandler struct {
	engine            *engine.Engine
	deployments       store.DeploymentStore
	deploymentLogs    store.DeploymentLogStore
	deploymentMetrics store.DeploymentMetricStore
	limits            config.ServerEngineLimitConfig
}

func NewHandler(
	engine *engine.Engine,
	deployments store.DeploymentStore,
	deploymentLogs store.DeploymentLogStore,
	deploymentMetrics store.DeploymentMetricStore,
	limits config.ServerEngineLimitConfig,
) *DeploymentHandler {
	return &DeploymentHandler{
		engine:            engine,
		deployments:       deployments,
		deploymentLogs:    deploymentLogs,
		deploymentMetrics: deploymentMetrics,
		limits:            limits,
	}
}

func (h *DeploymentHandler) HandleDeploymentCreate(c *fiber.Ctx, req wire.DeploymentCreateRequest) error {
	guildID := c.Params("guildID")

	module, err := module.New(c.Context(), req.WasmBytes, module.ModuleConfig{
		MemoryPagesLimit:   h.limits.MaxMemoryPages,
		TotalTimeLimit:     time.Duration(h.limits.MaxTotalTime) * time.Millisecond,
		ExecutionTimeLimit: time.Duration(h.limits.MaxExecutionTime) * time.Millisecond,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create module: %w", err)
	}

	manifest, err := module.GetManifest(c.Context())
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}

	deployment, err := h.deployments.UpsertDeployment(c.Context(), model.Deployment{
		ID:              util.UniqueID(),
		Name:            req.Name,
		Key:             req.Key,
		Description:     req.Description,
		GuildID:         guildID,
		PluginVersionID: req.PluginVersionID,
		WasmBytes:       req.WasmBytes,
		Manifest:        *manifest,
		Config:          req.Config,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	return c.JSON(wire.DeploymentCreateResponse{
		Success: true,
		Data:    wire.DeploymentToWire(deployment),
	})
}

func (h *DeploymentHandler) HandleDeploymentListForGuild(c *fiber.Ctx) error {
	deployments, err := h.deployments.GetDeploymentsForGuild(c.Context(), c.Params("guildID"))
	if err != nil {
		return err
	}

	res := make([]wire.Deployment, len(deployments))
	for i, deployment := range deployments {
		res[i] = wire.DeploymentToWire(&deployment)
	}

	return c.JSON(wire.DeploymentListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *DeploymentHandler) HandleDeploymentGet(c *fiber.Ctx) error {
	deployment, err := h.deployments.GetDeployment(c.Context(), c.Params("deploymentID"), c.Params("guildID"))
	if err != nil {
		if err == store.ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "unknown_deployment", "Deployment not found")
		}
		return err
	}

	return c.JSON(wire.DeploymentGetResponse{
		Success: true,
		Data:    wire.DeploymentToWire(deployment),
	})
}

func (h *DeploymentHandler) HandleDeploymentDelete(c *fiber.Ctx) error {
	err := h.deployments.DeleteDeployment(c.Context(), c.Params("deploymentID"), c.Params("guildID"))
	if err != nil {
		if err == store.ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "unknown_deployment", "Deployment not found")
		}
		return err
	}

	return c.JSON(wire.DeploymentDeleteResponse{
		Success: true,
	})
}
