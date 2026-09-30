package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kitecloud/kite/kite-support/internal/bot"
	"github.com/kitecloud/kite/kite-support/internal/config"
	"github.com/kitecloud/kite/kite-support/internal/embedded"
	"github.com/kitecloud/kite/kite-support/internal/llm"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/urfave/cli/v2"
)

var botCMD = cli.Command{
	Name:  "bot",
	Usage: "Run the Discord support bot.",
	Action: func(c *cli.Context) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if cfg.Discord.Token == "" {
			return fmt.Errorf("discord.token is required")
		}
		if cfg.OpenAI.APIKey == "" {
			return fmt.Errorf("openai.api_key is required")
		}

		k := embedded.Knowledge()
		if k == "" {
			return fmt.Errorf("no knowledge embedded in this binary, run `kite-support index` then rebuild")
		}

		oa := openai.NewClient(option.WithAPIKey(cfg.OpenAI.APIKey))
		llmClient := llm.New(&oa, llm.Config{
			Model:           cfg.OpenAI.Model,
			ReasoningEffort: cfg.OpenAI.ReasoningEffort,
			MaxOutputTokens: cfg.OpenAI.MaxOutputTokens,
		})

		b, err := bot.New(cfg, k, llmClient)
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		return b.Run(ctx)
	},
}
