package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kitecloud/kite/kite-support/internal/config"
	"github.com/kitecloud/kite/kite-support/internal/knowledge"
	"github.com/urfave/cli/v2"
)

var indexCMD = cli.Command{
	Name:  "index",
	Usage: "Build the knowledge from the docs and the block catalog.",
	Action: func(c *cli.Context) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		k, err := knowledge.Build(cfg.Knowledge.DocsPath, cfg.Knowledge.CatalogPath, cfg.Knowledge.DocsBaseURL)
		if err != nil {
			return err
		}
		if err := os.WriteFile(cfg.Knowledge.Path, []byte(k), 0o644); err != nil {
			return fmt.Errorf("write knowledge: %w", err)
		}

		slog.Info("knowledge written", "path", cfg.Knowledge.Path, "bytes", len(k))
		return nil
	},
}
