package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/responses"
	"github.com/openai/openai-go/v2/shared"
)

const (
	IntentQuestion   = "question"
	IntentBug        = "bug"
	IntentSuggestion = "suggestion"
)

var leakedFields = regexp.MustCompile(`\n\s*"?(intent|links)"?\s*:`)

type Config struct {
	Model           string
	ReasoningEffort string
	MaxOutputTokens int
}

type Client struct {
	openai *openai.Client
	config Config
}

func New(o *openai.Client, config Config) *Client {
	return &Client{openai: o, config: config}
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

// Answer answers the question from knowledge, which the model gets in full
// every time. It comes first and rarely changes, so the provider caches it.
func (c *Client) Answer(ctx context.Context, knowledge string, history []Turn, question, userID string) (*Answer, error) {
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
		Instructions:    openai.String(systemPrompt + "\n\n" + knowledge),
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: input},
		MaxOutputTokens: openai.Int(int64(c.config.MaxOutputTokens)),
		Reasoning: shared.ReasoningParam{
			Effort: shared.ReasoningEffort(c.config.ReasoningEffort),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "support_answer",
					Schema: outputSchema,
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

	var answer Answer
	if err := json.Unmarshal([]byte(resp.OutputText()), &answer); err != nil {
		return nil, fmt.Errorf("parse answer: %w", err)
	}
	// The model sometimes repeats the other fields at the end of the answer.
	if loc := leakedFields.FindStringIndex(answer.Text); loc != nil {
		answer.Text = answer.Text[:loc[0]]
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
