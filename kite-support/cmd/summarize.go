package cmd

import (
	"fmt"
	"log/slog"

	"github.com/kitecloud/kite/kite-support/internal/config"
	"github.com/kitecloud/kite/kite-support/internal/index"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/urfave/cli/v2"
)

var summarizeCMD = cli.Command{
	Name:  "summarize",
	Usage: "Regenerate concepts.md from the kite-service codebase.",
	Action: func(c *cli.Context) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if cfg.OpenAI.APIKey == "" {
			return fmt.Errorf("openai.api_key is required")
		}

		client := openai.NewClient(option.WithAPIKey(cfg.OpenAI.APIKey))
		model := cfg.OpenAI.SummaryModel
		if model == "" {
			model = "gpt-4o"
		}

		slog.Info("summarizing codebase", "service_path", cfg.Index.ServicePath, "out", cfg.Index.ConceptsPath, "model", model)
		if err := index.SummarizeCodebase(c.Context, &client, model, cfg.Index.ServicePath, cfg.Index.ConceptsPath); err != nil {
			return err
		}
		slog.Info("concepts.md written", "path", cfg.Index.ConceptsPath)
		return nil
	},
}
