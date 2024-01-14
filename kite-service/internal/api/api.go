package api

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres"
	"github.com/merlinfuchs/kite/kite-service/internal/handler/deployment"
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

	api.app.Post("/api/v1/deployments", helpers.WithRequestBody(deploymentHandler.CreateDeployment))
}

func (api *API) Serve(host string, port int) error {
	return api.app.Listen(fmt.Sprintf("%s:%d", host, port))
}

func (api *API) Shutdown() error {
	return api.app.Shutdown()
}
