package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kitecloud/kite/kite-support/internal/config"
	"github.com/kitecloud/kite/kite-support/internal/index"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/urfave/cli/v2"
)

const embeddingDim = 1536

var indexCMD = cli.Command{
	Name:  "index",
	Usage: "Build the documentation vector index.",
	Action: func(c *cli.Context) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if cfg.OpenAI.APIKey == "" {
			return fmt.Errorf("openai.api_key is required")
		}

		docs, err := index.WalkDocs(cfg.Index.DocsPath, cfg.Index.DocsBaseURL)
		if err != nil {
			return fmt.Errorf("walk docs: %w", err)
		}
		concepts, err := index.WalkConcepts(cfg.Index.ConceptsPath)
		if err != nil {
			return fmt.Errorf("walk concepts: %w", err)
		}
		all := append(docs, concepts...)
		slog.Info("documents loaded", "docs", len(docs), "concepts", len(concepts))

		chunks := index.BuildChunks(all, cfg.Index.ChunkSize, cfg.Index.ChunkOverlap)
		slog.Info("chunks built", "count", len(chunks))
		if len(chunks) == 0 {
			return fmt.Errorf("no chunks produced — check docs_path")
		}

		client := openai.NewClient(option.WithAPIKey(cfg.OpenAI.APIKey))
		embedder := index.NewEmbedder(&client, cfg.OpenAI.EmbeddingModel)

		store := index.NewStore()
		ctx := c.Context
		batchSize := 64
		for i := 0; i < len(chunks); i += batchSize {
			end := i + batchSize
			if end > len(chunks) {
				end = len(chunks)
			}
			batch := chunks[i:end]
			texts := make([]string, len(batch))
			for j, ch := range batch {
				texts[j] = ch.Heading + "\n\n" + ch.Content
			}
			vecs, err := embedder.EmbedBatch(ctx, texts)
			if err != nil {
				return fmt.Errorf("embed batch %d: %w", i/batchSize, err)
			}
			if err := store.Insert(batch, vecs); err != nil {
				return fmt.Errorf("insert batch %d: %w", i/batchSize, err)
			}
			slog.Info("embedded", "done", end, "total", len(chunks))
		}

		f, err := os.Create(cfg.Index.DBPath)
		if err != nil {
			return fmt.Errorf("create index file: %w", err)
		}
		defer f.Close()
		if err := store.Encode(f); err != nil {
			return fmt.Errorf("encode index: %w", err)
		}

		slog.Info("index complete", "path", cfg.Index.DBPath, "chunks", store.Count())
		return nil
	},
}
