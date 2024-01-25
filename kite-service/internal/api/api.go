package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
	kiteapp "github.com/merlinfuchs/kite/kite-app"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres"
	"github.com/merlinfuchs/kite/kite-service/internal/handler/compile"
	"github.com/merlinfuchs/kite/kite-service/internal/handler/deployment"
	guild "github.com/merlinfuchs/kite/kite-service/internal/handler/guilld"
	quickaccess "github.com/merlinfuchs/kite/kite-service/internal/handler/quick_access"
	"github.com/merlinfuchs/kite/kite-service/internal/handler/workspace"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type API struct {
	app *fiber.App
}

func New() *API {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var e *wire.Error
			if errors.As(err, &e) {
				return c.Status(e.Status).JSON(e)
			} else {
				slog.With(logattr.Error(err)).Error("Unhandled error in rest endpoint")
				return c.Status(fiber.StatusInternalServerError).JSON(wire.Error{
					Status:  fiber.StatusInternalServerError,
					Code:    "internal_server_error",
					Message: err.Error(),
				})
			}
		},
		BodyLimit: 1024 * 1024 * 8, // 8 MB
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	api := &API{
		app: app,
	}

	return api
}

func (api *API) RegisterHandlers(engine *engine.PluginEngine, pg *postgres.Client) {
	deploymentHandler := deployment.NewHandler(engine, pg)
	api.app.Get("/api/v1/guilds/:guildID/deployments", deploymentHandler.HandleDeploymentListForGuild)
	api.app.Post("/api/v1/guilds/:guildID/deployments", helpers.WithRequestBody(deploymentHandler.HandleDeploymentCreate))
	api.app.Delete("/api/v1/guilds/:guildID/deployments/:deploymentID", deploymentHandler.HandleDeploymentDelete)
	api.app.Get("/api/v1/guilds/:guildID/deployments/:deploymentID/logs", deploymentHandler.HandleDeploymentLogEntryList)
	api.app.Get("/api/v1/guilds/:guildID/deployments/:deploymentID/logs/summary", deploymentHandler.HandleDeploymentLogSummaryGet)

	guildHandler := guild.NewHandler(engine, pg)
	api.app.Get("/api/v1/guilds", guildHandler.HandleGuildList)
	api.app.Get("/api/v1/guilds/:guildID", guildHandler.HandleGuildGet)

	workspaceHandler := workspace.NewHandler(pg)
	api.app.Post("/api/v1/guilds/:guildID/workspaces", helpers.WithRequestBody(workspaceHandler.HandleWorkspaceCreate))
	api.app.Put("/api/v1/guilds/:guildID/workspaces/:workspaceID", helpers.WithRequestBody(workspaceHandler.HandleWorkspaceUpdate))
	api.app.Get("/api/v1/guilds/:guildID/workspaces/:workspaceID", workspaceHandler.HandleWorkspaceGetForGuild)
	api.app.Get("/api/v1/guilds/:guildID/workspaces", workspaceHandler.HandleWorkspaceListForGuild)
	api.app.Delete("/api/v1/guilds/:guildID/workspaces/:workspaceID", workspaceHandler.HandleWorkspaceDelete)

	quickAccessHandler := quickaccess.NewHandler(pg)
	api.app.Get("/api/v1/guilds/:guildID/quick-access", quickAccessHandler.HandleQuickAccessItemList)

	compileHandler := compile.NewHandler()
	api.app.Post("/api/v1/compile/js", helpers.WithRequestBody(compileHandler.HandleCompileJS))

	// Serve statix files
	api.app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(kiteapp.OutFS),
		Browse:       false,
		NotFoundFile: "out/index.html",
		PathPrefix:   "/out",
	}))
}

func (api *API) Serve(host string, port int) error {
	return api.app.Listen(fmt.Sprintf("%s:%d", host, port))
}

func (api *API) Shutdown() error {
	return api.app.Shutdown()
}
