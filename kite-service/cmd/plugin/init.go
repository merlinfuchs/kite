package plugin

import (
	"fmt"
	"os"

	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/urfave/cli/v2"
)

func initCMD() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize a new Kite plugin in the current directory",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "type",
				Usage:    "Type of the plugin. (Go or JS)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "Name of the plugin",
			},
			&cli.StringFlag{
				Name:  "description",
				Usage: "Description of the plugin",
			},
		},
		Action: func(c *cli.Context) error {
			basePath := c.String("path")
			typ := c.String("type")
			name := c.String("name")
			description := c.String("description")

			return runInit(basePath, typ, name, description)
		},
	}
}

func runInit(basePath, typ, name, description string) error {
	if _, err := os.Stat(basePath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(basePath, 0755); err != nil {
				return fmt.Errorf("Failed to create directory: %v", err)
			}
		} else {
			return fmt.Errorf("Failed to stat directory: %v", err)
		}
	}

	if config.ConfigExists(basePath) {
		return fmt.Errorf("A Kite config already exists in this directory")
	}

	cfg, err := config.DefaultPluginConifg()
	if err != nil {
		return fmt.Errorf("Failed to load default config: %v", err)
	}

	if name != "" {
		cfg.Name = name
	}

	if description != "" {
		cfg.Description = description
	}

	cfg.Type = typ

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("Invalid plugin config: %v", err)
	}

	err = config.WritePluginConfig(basePath, cfg)
	if err != nil {
		return fmt.Errorf("Failed to write config: %v", err)
	}

	return nil
}
