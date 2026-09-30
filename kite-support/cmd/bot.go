package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kitecloud/kite/kite-support/internal/bot"
	"github.com/kitecloud/kite/kite-support/internal/config"
	"github.com/kitecloud/kite/kite-support/internal/embedded"
	"github.com/kitecloud/kite/kite-support/internal/index"
	"github.com/kitecloud/kite/kite-support/internal/llm"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
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

		if len(embedded.Index) == 0 {
			return fmt.Errorf("no index embedded in this binary — run `kite-support index` then rebuild")
		}
		store := index.NewStore()
		if err := store.Decode(bytes.NewReader(embedded.Index)); err != nil {
			return fmt.Errorf("decode embedded index: %w", err)
		}
		if store.Count() == 0 {
			return fmt.Errorf("embedded index is empty — run `kite-support index` then rebuild")
		}

		oa := openai.NewClient(option.WithAPIKey(cfg.OpenAI.APIKey))
		embedder := index.NewEmbedder(&oa, cfg.OpenAI.EmbeddingModel)
		llmClient := llm.New(&oa, cfg.OpenAI.ChatModel)

		b, err := bot.New(cfg, store, embedder, llmClient)
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		return b.Run(ctx)
	},
}
