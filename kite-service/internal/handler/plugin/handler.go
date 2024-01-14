package plugin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type PluginHandler struct {
	engine  *engine.PluginEngine
	plugins store.PluginStore
}

func NewHandler(engine *engine.PluginEngine, plugins store.PluginStore) *PluginHandler {
	return &PluginHandler{
		engine:  engine,
		plugins: plugins,
	}
}

func (h *PluginHandler) CreatePlugin(c *fiber.Ctx, req wire.PluginCreateRequest) error {
	return nil
}

func (h *PluginHandler) CreatePluginDeployment(c *fiber.Ctx, req wire.PluginDeploymentCreateRequest) error {
	return nil
}

func (h *PluginHandler) OverridePluginDeployments(c *fiber.Ctx, req wire.PluginDeploymentsOverrideRequest) error {
	return nil
}
