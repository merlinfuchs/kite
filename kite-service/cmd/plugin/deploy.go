package plugin

import (
	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/urfave/cli/v2"
)

func deployCMD() *cli.Command {
	return &cli.Command{
		Name:  "deploy",
		Usage: "Deploy a plugin to a specific guild on a Kite server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild_id",
				Usage:    "Guild ID to deploy the plugin to",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "config",
				Usage: "JSON config for the plugin",
			},
		},
		Action: func(c *cli.Context) error {
			basePath := c.String("path")
			guildID := c.String("guild_id")
			configPath := c.String("config")

			cfg, err := config.LoadPluginConfig(basePath)
			if err != nil {
				return err
			}

			return runDeploy(basePath, guildID, configPath, cfg)
		},
	}
}

func runDeploy(basePath, guildID, configPath string, cfg *config.PluginConfig) error {
	return nil
}
