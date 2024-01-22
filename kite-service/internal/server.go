package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/api"
	"github.com/merlinfuchs/kite/kite-service/internal/bot"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres"
	"github.com/merlinfuchs/kite/kite-service/internal/deployments"
	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
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

	env := host.NewEnv(bot)
	e := engine.New(env)

	manager := deployments.NewManager(pg, e)
	manager.Start()

	bot.Engine = e

	for _, pl := range cfg.StaticPlugins {
		pluginCFG, err := config.LoadPluginConfig(pl.Path)
		if err != nil {
			return fmt.Errorf("failed to load plugin config: %w", err)
		}

		wasm, err := os.ReadFile(filepath.Join(pl.Path, pluginCFG.Build.Out))
		if err != nil {
			return fmt.Errorf("failed to read plugin wasm: %w", err)
		}

		userConfig := make(map[string]string, len(pluginCFG.DefaultConfig)+len(pl.Config))
		for k, v := range pluginCFG.DefaultConfig {
			userConfig[k] = v
		}
		for k, v := range pl.Config {
			userConfig[k] = v
		}

		commands := make([]plugin.ManifestCommand, len(pluginCFG.Commands))
		for i, cmd := range pluginCFG.Commands {
			commands[i] = plugin.ManifestCommand{
				Name:        cmd.Name,
				Description: cmd.Description,
				// TODO: Options:     cmd.Options,
			}
		}

		manifest := plugin.Manifest{
			ID:       "abc",
			Events:   pluginCFG.Events,
			Commands: commands,
		}
		config := plugin.PluginConfig{
			MemoryPagesLimit:   32,
			UserConfig:         userConfig,
			TotalTimeLimit:     time.Second * 10,
			ExecutionTimeLimit: time.Millisecond * 20,
		}

		deployment := engine.NewDeployment(wasm, manifest, config)
		err = e.LoadStaticDeployment(deployment)
		if err != nil {
			return fmt.Errorf("failed to load plugin: %w", err)
		}
	}

	commands := e.GlobalCommands()
	_, err = bot.Session.ApplicationCommandBulkOverwrite(cfg.Discord.ClientID, "", commands)
	if err != nil {
		return fmt.Errorf("failed to overwrite commands: %w", err)
	}

	err = bot.Start()
	if err != nil {
		return fmt.Errorf("failed to start discord bot: %w", err)
	}

	api := api.New()

	api.RegisterHandlers(e, pg)

	return api.Serve(cfg.Host, cfg.Port)
}
