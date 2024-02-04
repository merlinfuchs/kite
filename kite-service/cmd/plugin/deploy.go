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
				Value: "https://api.kite.onl",
			},
		},
		Action: func(c *cli.Context) error {
			basePath := c.Args().Get(0)
			if basePath == "" {
				basePath = "."
			}
			guildID := c.String("guild_id")
			rawUserConfig := c.String("config")
			server := c.String("server")

			cfg, err := config.LoadworkspaceConfig(basePath)
			if err != nil {
				return err
			}

			globalCFG, err := config.LoadGlobalConfig()
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

			return runDeploy(basePath, guildID, userConfig, serverURL, cfg, globalCFG)
		},
	}
}

func runDeploy(
	basePath string,
	guildID string,
	userConfig map[string]string,
	serverURL *url.URL,
	cfg *config.WorkspaceConfig,
	globalCFG *config.GlobalConfig,
) error {
	session := globalCFG.GetSessionForServer(serverURL.String())
	if session == nil {
		return fmt.Errorf("No session for server %s, login first!", serverURL.String())
	}

	serverURL.Path = path.Join(serverURL.Path, "v1/guilds", guildID, "deployments")

	wasmPath := filepath.Join(basePath, cfg.Module.Build.Out)
	wasm, err := os.ReadFile(wasmPath)
	if err != nil {
		return err
	}

	rawBody, err := json.Marshal(wire.DeploymentCreateRequest{
		Key:         cfg.Deployment.Key,
		Name:        cfg.Deployment.Name,
		Description: cfg.Deployment.Description,
		WasmBytes:   base64.StdEncoding.EncodeToString(wasm),
		// TODO: Config:
	})
	if err != nil {
		return fmt.Errorf("Failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", serverURL.String(), bytes.NewBuffer(rawBody))
	if err != nil {
		return fmt.Errorf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", session.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to deploy plugin: %v", err)
	}

	fmt.Println(resp.Status)

	return nil
}
