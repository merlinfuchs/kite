package deployment

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/util"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type DeploymentHandler struct {
	engine            *engine.PluginEngine
	deployments       store.DeploymentStore
	deploymentLogs    store.DeploymentLogStore
	deploymentMetrics store.DeploymentMetricStore
}

func NewHandler(engine *engine.PluginEngine, deployments store.DeploymentStore, deploymentLogs store.DeploymentLogStore, deploymentMetrics store.DeploymentMetricStore) *DeploymentHandler {
	return &DeploymentHandler{
		engine:            engine,
		deployments:       deployments,
		deploymentLogs:    deploymentLogs,
		deploymentMetrics: deploymentMetrics,
	}
}

func (h *DeploymentHandler) HandleDeploymentCreate(c *fiber.Ctx, req wire.DeploymentCreateRequest) error {
	wasmBytes, err := base64.StdEncoding.DecodeString(req.WasmBytes)
	if err != nil {
		return fmt.Errorf("failed to decode wasm bytes: %w", err)
	}

	guildID := c.Params("guildID")

	deployment, err := h.deployments.UpsertDeployment(c.Context(), model.Deployment{
		ID:                    util.UniqueID(),
		Name:                  req.Name,
		Key:                   req.Key,
		Description:           req.Description,
		GuildID:               guildID,
		PluginVersionID:       req.PluginVersionID,
		WasmBytes:             wasmBytes,
		ManifestDefaultConfig: req.ManifestDefaultConfig,
		ManifestEvents:        req.ManifestEvents,
		ManifestCommands:      req.ManifestCommands,
		Config:                req.Config,
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
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
