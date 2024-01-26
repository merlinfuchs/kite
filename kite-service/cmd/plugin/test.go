package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
	"github.com/urfave/cli/v2"
)

func testCMD() *cli.Command {
	return &cli.Command{
		Name:  "test",
		Usage: "Test the Kite plugin in the current directory",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "build",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
			},
		},
		Action: func(c *cli.Context) error {
			basePath := c.String("path")

			cfg, err := config.LoadPluginConfig(basePath)
			if err != nil {
				return err
			}

			return runTest(basePath, c.Bool("build"), c.Bool("debug"), cfg)
		},
	}
}

func runTest(basePath string, build bool, debug bool, cfg *config.PluginConfig) error {
	if build {
		if err := runBuild(basePath, debug, cfg); err != nil {
			return err
		}
	}

	wasmPath := filepath.Join(basePath, cfg.Build.Out)
	wasm, err := os.ReadFile(wasmPath)
	if err != nil {
		return err
	}

	config := plugin.PluginConfig{
		MemoryPagesLimit:   32,
		TotalTimeLimit:     time.Microsecond,
		ExecutionTimeLimit: time.Microsecond,
		UserConfig:         cfg.DefaultConfig,
	}

	manifest := plugin.Manifest{}

	ctx := context.Background()
	plugin, err := plugin.New(ctx, wasm, manifest, config, host.HostEnvironment{})
	if err != nil {
		return err
	}

	res, err := plugin.Handle(ctx, &event.Event{
		Type: event.DiscordMessageCreate,
		Data: &dismodel.MessageCreateEvent{
			Content: "test",
		},
	})
	if err != nil {
		return err
	} else {
		fmt.Println("Execution duration: ", res.ExecutionDuration)
		fmt.Println("Total duration: ", res.TotalDuration)
	}

	return nil
}
