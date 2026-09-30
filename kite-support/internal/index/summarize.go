package index

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/openai/openai-go"
)

const summarizeSystemPrompt = `You are a technical writer helping non-technical Discord users understand Kite, an open-source no-code platform for building Discord bots.

You will receive Go source code from the Kite backend. Your job is to produce a single Markdown document that explains the user-facing concepts of Kite (commands, message templates, event listeners, flows, plugins, variables, apps) in plain language.

Rules:
- Write for Discord users who do not code.
- No code snippets, no Go syntax, no API references.
- Use short sections with H2 headings per concept.
- Explain WHAT each concept is and WHEN a user would touch it.
- If something is implementation detail (database, caching, internal types), skip it.
- Be accurate; if the code does not show something, do not invent it.
- Aim for ~1500-3000 words total.`

func SummarizeCodebase(ctx context.Context, client *openai.Client, model, servicePath, outPath string) error {
	if servicePath == "" {
		return fmt.Errorf("service_path is empty")
	}
	bundle, err := bundleSource(servicePath)
	if err != nil {
		return err
	}

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(summarizeSystemPrompt),
			openai.UserMessage("Source code follows. Produce concepts.md.\n\n" + bundle),
		},
	})
	if err != nil {
		return fmt.Errorf("chat: %w", err)
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("no completion returned")
	}

	body := resp.Choices[0].Message.Content
	return os.WriteFile(outPath, []byte(body), 0o644)
}

func bundleSource(root string) (string, error) {
	wantDirs := []string{"cmd", "internal/core", "pkg/flow", "pkg/plugin", "pkg/message"}
	wantFiles := []string{"README.md"}

	var b strings.Builder
	addFile := func(path, rel string) error {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "\n\n=== FILE: %s ===\n%s\n", rel, raw)
		return nil
	}

	for _, f := range wantFiles {
		p := filepath.Join(root, f)
		if _, err := os.Stat(p); err == nil {
			if err := addFile(p, f); err != nil {
				return "", err
			}
		}
	}

	for _, d := range wantDirs {
		dirPath := filepath.Join(root, d)
		_ = filepath.WalkDir(dirPath, func(path string, de fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if de.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			return addFile(path, rel)
		})
	}

	const maxBytes = 600_000
	out := b.String()
	if len(out) > maxBytes {
		out = out[:maxBytes] + "\n\n[truncated]"
	}
	return out, nil
}
