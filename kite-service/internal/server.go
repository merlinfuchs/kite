package internal

import (
	"fmt"

	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/api"
	"github.com/merlinfuchs/kite/kite-service/internal/api/access"
	"github.com/merlinfuchs/kite/kite-service/internal/bot"
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

	envStores := host.NewHostEnvironmentStores(bot, pg, pg, pg, pg)
	manager, err := deployments.NewManager(pg, e, envStores)
	if err != nil {
		return fmt.Errorf("failed to create deployment manager: %w", err)
	}
	manager.Start()

	bot.Engine = e

	err = bot.Start()
	if err != nil {
		return fmt.Errorf("failed to start discord bot: %w", err)
	}

	api := api.New()

	accessManager := access.New(bot.State)
	api.RegisterHandlers(e, pg, accessManager, cfg)

	return api.Serve(cfg.Host, cfg.Port)
}
