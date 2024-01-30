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
	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/api/access"
	"github.com/merlinfuchs/kite/kite-service/internal/api/handler/auth"
	"github.com/merlinfuchs/kite/kite-service/internal/api/handler/compile"
	"github.com/merlinfuchs/kite/kite-service/internal/api/handler/deployment"
	guild "github.com/merlinfuchs/kite/kite-service/internal/api/handler/guilld"
	kvstorage "github.com/merlinfuchs/kite/kite-service/internal/api/handler/kv_storage"
	quickaccess "github.com/merlinfuchs/kite/kite-service/internal/api/handler/quick_access"
	"github.com/merlinfuchs/kite/kite-service/internal/api/handler/user"
	"github.com/merlinfuchs/kite/kite-service/internal/api/handler/workspace"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres"
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

func (api *API) RegisterHandlers(engine *engine.Engine, pg *postgres.Client, accessManager *access.AccessManager, cfg *config.ServerConfig) {
	sessionManager := session.New(pg)
	sessionMiddleware := session.NewMiddleware(sessionManager)
	accessMiddleware := access.NewMiddleware(accessManager)

	apiGroup := api.app.Group("/api/v1")

	authHandler := auth.New(sessionManager, pg, cfg)
	apiGroup.Get("/auth/redirect", authHandler.HandleAuthRedirect)
	apiGroup.Get("/auth/callback", authHandler.HandleAuthCallback)
	apiGroup.Get("/auth/logout", authHandler.HandleAuthLogout)
	apiGroup.Post("/auth/cli/start", authHandler.HandleAuthCLIStart)
	apiGroup.Get("/auth/cli/redirect", authHandler.HandleAuthCLIRedirect)
	apiGroup.Get("/auth/cli/callback", authHandler.HandleAuthCLICallback)
	apiGroup.Get("/auth/cli/check", authHandler.HandleAuthCLICheck)

	userGroup := apiGroup.Group("/users").Use(sessionMiddleware.SessionRequired())
	userHandler := user.NewHandler(pg)
	userGroup.Get("/:userID", userHandler.HandleUserGet)

	guildsGroup := apiGroup.Group("/guilds").Use(sessionMiddleware.SessionRequired())
	guildGroup := guildsGroup.Group("/:guildID").Use(accessMiddleware.GuildAccessRequired())

	guildHandler := guild.NewHandler(pg, accessManager)
	guildsGroup.Get("/", guildHandler.HandleGuildList)
	guildGroup.Get("/", guildHandler.HandleGuildGet)

	deploymentHandler := deployment.NewHandler(engine, pg, pg, pg)
	guildGroup.Get("/deployments", deploymentHandler.HandleDeploymentListForGuild)
	guildGroup.Post("/deployments", helpers.WithRequestBody(deploymentHandler.HandleDeploymentCreate))
	guildGroup.Get("/deployments/:deploymentID", deploymentHandler.HandleDeploymentGet)
	guildGroup.Delete("/deployments/:deploymentID", deploymentHandler.HandleDeploymentDelete)
	guildGroup.Get("/deployments/:deploymentID/logs", deploymentHandler.HandleDeploymentLogEntryList)
	guildGroup.Get("/deployments/:deploymentID/logs/summary", deploymentHandler.HandleDeploymentLogSummaryGet)
	guildGroup.Get("/deployments/:deploymentID/metrics/events", deploymentHandler.HandleDeploymentEventMetricsList)
	guildGroup.Get("/deployments/:deploymentID/metrics/calls", deploymentHandler.HandleDeploymentCallMetricsList)
	guildGroup.Get("/deployments/metrics/events", deploymentHandler.HandleDeploymentsEventMetricsList)
	guildGroup.Get("/deployments/metrics/calls", deploymentHandler.HandleDeploymentsCallMetricsList)

	workspaceHandler := workspace.NewHandler(pg)
	guildGroup.Post("/workspaces", helpers.WithRequestBody(workspaceHandler.HandleWorkspaceCreate))
	guildGroup.Put("/workspaces/:workspaceID", helpers.WithRequestBody(workspaceHandler.HandleWorkspaceUpdate))
	guildGroup.Get("/workspaces/:workspaceID", workspaceHandler.HandleWorkspaceGetForGuild)
	guildGroup.Get("/workspaces", workspaceHandler.HandleWorkspaceListForGuild)
	guildGroup.Delete("/workspaces/:workspaceID", workspaceHandler.HandleWorkspaceDelete)

	kvStorageHandler := kvstorage.NewHandler(pg)
	guildGroup.Get("/kv-storage/namespaces", kvStorageHandler.HandleKVStorageNamespaceList)
	guildGroup.Get("/kv-storage/namespaces/:namespace/keys", kvStorageHandler.HandleKVStorageNamespaceKeyList)

	quickAccessHandler := quickaccess.NewHandler(pg)
	guildGroup.Get("/quick-access", quickAccessHandler.HandleQuickAccessItemList)

	compileHandler := compile.NewHandler()
	apiGroup.Post("/compile/js", helpers.WithRequestBody(compileHandler.HandleCompileJS))

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
