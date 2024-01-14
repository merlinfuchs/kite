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
	engine      *engine.PluginEngine
	deployments store.DeploymentStore
}

func NewHandler(engine *engine.PluginEngine, deployments store.DeploymentStore) *DeploymentHandler {
	return &DeploymentHandler{
		engine:      engine,
		deployments: deployments,
	}
}

func (h *DeploymentHandler) CreateDeployment(c *fiber.Ctx, req wire.DeploymentCreateRequest) error {
	wasmBytes, err := base64.StdEncoding.DecodeString(req.WasmBytes)
	if err != nil {
		return fmt.Errorf("failed to decode wasm bytes: %w", err)
	}

	deployment, err := h.deployments.UpsertDeployment(c.Context(), model.Deployment{
		ID:                    util.UniqueID(),
		Name:                  req.Name,
		Key:                   req.Key,
		Description:           req.Description,
		GuildID:               req.GuildID,
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
