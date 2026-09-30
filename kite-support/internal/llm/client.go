package llm

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kitecloud/kite/kite-support/internal/index"
	"github.com/openai/openai-go"
)

var bareURL = regexp.MustCompile(`(?:^|[^<\w])(https?://[^\s<>)\]]+)`)
var intentTag = regexp.MustCompile(`(?m)^\s*\[(OK|BUG|SUGG)\]\s*$`)

const (
	IntentOK   = "OK"
	IntentBug  = "BUG"
	IntentSugg = "SUGG"
)

type Client struct {
	openai *openai.Client
	model  string
}

func New(o *openai.Client, model string) *Client {
	return &Client{openai: o, model: model}
}

func (c *Client) Answer(ctx context.Context, question string, hits []index.Hit) (string, string, error) {
	contextBlock := formatContext(hits)
	if strings.TrimSpace(contextBlock) == "" {
		contextBlock = "(no relevant documentation found)"
	}

	resp, err := c.openai.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(fmt.Sprintf("Documentation:\n%s\n\nQuestion: %s", contextBlock, question)),
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("chat: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", "", fmt.Errorf("no completion returned")
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	intent := IntentOK
	if m := intentTag.FindStringSubmatch(raw); m != nil {
		intent = m[1]
	}
	body := strings.TrimSpace(intentTag.ReplaceAllString(raw, ""))
	return suppressLinkEmbeds(body), intent, nil
}

func suppressLinkEmbeds(s string) string {
	return bareURL.ReplaceAllStringFunc(s, func(m string) string {
		idx := strings.Index(m, "http")
		return m[:idx] + "<" + m[idx:] + ">"
	})
}
