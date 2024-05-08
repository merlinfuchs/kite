package internal

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/merlinfuchs/kite/kite-service/internal/api"
	"github.com/merlinfuchs/kite/kite-service/internal/api/access"
	"github.com/merlinfuchs/kite/kite-service/internal/app"
	"github.com/merlinfuchs/kite/kite-service/internal/config"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres"
	"github.com/merlinfuchs/kite/kite-service/internal/deployments"
	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/internal/jobs"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
)

func RunServer(cfg *config.ServerConfig) error {
	pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Postgres))
	if err != nil {
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	jobs, err := jobs.NewClient(pg.DB, jobs.DefaultWorkers(pg, pg, pg), jobs.DefaultPeriodicJobs())
	if err != nil {
		return fmt.Errorf("failed to create jobs client: %w", err)
	}

	go func() {
		if err := jobs.Start(context.Background()); err != nil {
			slog.With(logattr.Error(err)).Error("Failed to start jobs client")
		}
	}()

	appManager := app.NewManager(pg, pg)

	e := engine.New()

	envStores := host.NewHostEnvironmentStores(pg, pg, pg, pg, appManager)
	manager, err := deployments.NewManager(pg, e, envStores, appManager, cfg.Engine.Limits)
	if err != nil {
		return fmt.Errorf("failed to create deployment manager: %w", err)
	}
	manager.Start()

	appManager.Engine = e

	go appManager.Run(context.Background())

	api := api.New(cfg)

	accessManager := access.New(pg)
	api.RegisterHandlers(e, pg, accessManager, cfg)

	return api.Serve(cfg.Host, cfg.Port)
}
