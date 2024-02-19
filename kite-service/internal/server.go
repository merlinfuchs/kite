package internal

import (
	"context"
	"fmt"

	"github.com/merlinfuchs/kite/kite-service/internal/api"
	"github.com/merlinfuchs/kite/kite-service/internal/api/access"
	"github.com/merlinfuchs/kite/kite-service/internal/bot"
	"github.com/merlinfuchs/kite/kite-service/internal/config"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres"
	"github.com/merlinfuchs/kite/kite-service/internal/deployments"
	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
)

func RunServer(cfg *config.ServerConfig) error {
	pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Postgres))
	if err != nil {
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	bot, err := bot.New(cfg.Discord.Token, pg)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	e := engine.New()

	envStores := host.NewHostEnvironmentStores(pg, pg, pg, pg, bot.State, bot.Client)
	manager, err := deployments.NewManager(pg, e, envStores, bot.Client, cfg.Engine.Limits)
	if err != nil {
		return fmt.Errorf("failed to create deployment manager: %w", err)
	}
	manager.Start()

	bot.Engine = e

	bot.Open(context.Background())

	api := api.New(cfg)

	accessManager := access.New(bot.State)
	api.RegisterHandlers(e, pg, accessManager, cfg)

	return api.Serve(cfg.Host, cfg.Port)
}
