package plugin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
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
				Value: "{}",
			},
			&cli.StringFlag{
				Name:  "server",
				Usage: "The Kite server to deploy to",
				Value: "http://localhost:8080",
			},
		},
		Action: func(c *cli.Context) error {
			basePath := c.String("path")
			guildID := c.String("guild_id")
			rawUserConfig := c.String("config")
			server := c.String("server")

			cfg, err := config.LoadPluginConfig(basePath)
			if err != nil {
				return err
			}

			userConfig := make(map[string]string)
			if err := json.Unmarshal([]byte(rawUserConfig), &userConfig); err != nil {
				return fmt.Errorf("Failed to parse user config: %v", err)
			}

			serverURL, err := url.Parse(server)
			if err != nil {
				return fmt.Errorf("Failed to parse server URL: %v", err)
			}

			return runDeploy(basePath, guildID, userConfig, serverURL, cfg)
		},
	}
}

func runDeploy(basePath string, guildID string, userConfig map[string]string, serverURL *url.URL, cfg *config.PluginConfig) error {
	serverURL.Path = path.Join(serverURL.Path, "api/v1/deployments")

	wasmPath := filepath.Join(basePath, cfg.Build.Out)
	wasm, err := os.ReadFile(wasmPath)
	if err != nil {
		return err
	}

	rawBody, err := json.Marshal(wire.DeploymentCreateRequest{
		Key:                   cfg.Key,
		Name:                  cfg.Name,
		Description:           cfg.Description,
		GuildID:               guildID,
		ManifestDefaultConfig: cfg.DefaultConfig,
		ManifestEvents:        cfg.Events,
		WasmBytes:             base64.StdEncoding.EncodeToString(wasm),
		// ManifestCommands:      cfg.Commands,
		Config: cfg.DefaultConfig,
	})
	if err != nil {
		return fmt.Errorf("Failed to marshal request body: %v", err)
	}

	resp, err := http.Post(serverURL.String(), "application/json", bytes.NewBuffer(rawBody))
	if err != nil {
		return fmt.Errorf("Failed to deploy plugin: %v", err)
	}

	fmt.Println(resp.Status)

	return nil
}
