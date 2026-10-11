package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/responses"
	"github.com/openai/openai-go/v2/shared"
)

const (
	IntentQuestion   = "question"
	IntentBug        = "bug"
	IntentSuggestion = "suggestion"
)

type Config struct {
	Model           string
	ReasoningEffort string
	MaxOutputTokens int
}

type Client struct {
	openai *openai.Client
	config Config
	schema map[string]any
}

// New creates a client whose answers only link to pageURLs.
func New(o *openai.Client, config Config, pageURLs []string) *Client {
	return &Client{openai: o, config: config, schema: outputSchema(pageURLs)}
}

// Turn is an earlier question and answer, so follow-ups can refer to them.
type Turn struct {
	Question string
	Answer   string
}

type Answer struct {
	Text   string   `json:"answer"`
	Intent string   `json:"intent"`
	Links  []string `json:"links"`
}

// Answer answers the question with instructions from Instructions.
func (c *Client) Answer(ctx context.Context, instructions string, history []Turn, question, userID string) (*Answer, error) {
	input := make(responses.ResponseInputParam, 0, len(history)*2+1)
	for _, t := range history {
		input = append(input,
			message(responses.EasyInputMessageRoleUser, t.Question),
			message(responses.EasyInputMessageRoleAssistant, t.Answer),
		)
	}
	input = append(input, message(responses.EasyInputMessageRoleUser, question))

	userHash := sha256.Sum256([]byte(userID))
	resp, err := c.openai.Responses.New(ctx, responses.ResponseNewParams{
		Model:           c.config.Model,
		Instructions:    openai.String(instructions),
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: input},
		MaxOutputTokens: openai.Int(int64(c.config.MaxOutputTokens)),
		Reasoning: shared.ReasoningParam{
			Effort: shared.ReasoningEffort(c.config.ReasoningEffort),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "support_answer",
					Schema: c.schema,
					Strict: openai.Bool(true),
				},
			},
		},
		PromptCacheKey: openai.String("kite-support"),
		// Lets OpenAI tell users apart for abuse detection.
		SafetyIdentifier: openai.String(hex.EncodeToString(userHash[:])),
	})
	if err != nil {
		return nil, fmt.Errorf("create response: %w", err)
	}

	slog.Info("answered question",
		"status", resp.Status,
		"input_tokens", resp.Usage.InputTokens,
		"cached_input_tokens", resp.Usage.InputTokensDetails.CachedTokens,
		"output_tokens", resp.Usage.OutputTokens,
	)
	if resp.Status != responses.ResponseStatusCompleted {
		return nil, fmt.Errorf("response %s: %s", resp.Status, resp.IncompleteDetails.Reason)
	}

	// A refusal has no output text.
	text := resp.OutputText()
	if text == "" {
		return &Answer{Text: "Sorry, I can't help with that. I can only answer questions about Kite.", Intent: IntentQuestion}, nil
	}
	var answer Answer
	if err := json.Unmarshal([]byte(text), &answer); err != nil {
		return nil, fmt.Errorf("parse answer: %w", err)
	}
	return &answer, nil
}

func message(role responses.EasyInputMessageRole, content string) responses.ResponseInputItemUnionParam {
	return responses.ResponseInputItemUnionParam{
		OfMessage: &responses.EasyInputMessageParam{
			Role: role,
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: openai.String(content),
			},
		},
	}
}
