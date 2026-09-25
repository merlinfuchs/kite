package flowai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/responses"
)

type CheckRequest struct {
	// Flow is the flow as serialized by the editor.
	Flow string
	// Prompt is the user's first message in the chat.
	Prompt string
	AppID  string
	UserID string
}

// CheckResponse says whether a prompt is ready for the flow AI, or suggests a
// clearer one with fields for what's missing.
type CheckResponse struct {
	Verdict         string       `json:"verdict"`
	Message         string       `json:"message"`
	SuggestedPrompt string       `json:"suggested_prompt"`
	Fields          []CheckField `json:"fields"`
}

type CheckField struct {
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Options     []string `json:"options"`
	Default     string   `json:"default"`
}

// Check asks a cheaper model whether the user's first prompt has what the
// flow AI needs, before it uses up one of the user's prompts.
func (a *Assistant) Check(ctx context.Context, req CheckRequest) (*CheckResponse, error) {
	resp, _, err := a.call(ctx, call{
		model:        a.config.Check,
		instructions: checkInstructions,
		input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(fmt.Sprintf("Current flow:\n%s\n\nThe user's request:\n%s", req.Flow, req.Prompt)),
		},
		schemaName: "prompt_check",
		schema:     checkSchema,
		cacheKey:   "kite-flow-ai-check",
		userID:     req.UserID,
		logMessage: "Flow AI check",
		logAttrs:   []any{slog.String("app_id", req.AppID)},
	})
	if err != nil {
		return nil, err
	}
	if resp.Status != responses.ResponseStatusCompleted {
		return nil, fmt.Errorf("response is %s", resp.Status)
	}

	var res CheckResponse
	if err := json.Unmarshal([]byte(resp.OutputText()), &res); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &res, nil
}

var checkSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []string{"verdict", "message", "suggested_prompt", "fields"},
	"properties": map[string]any{
		"verdict": map[string]any{
			"type": "string",
			"enum": []string{"send", "clarify"},
		},
		"message": map[string]any{
			"type":        "string",
			"description": "clarify: what's missing, in one or two short sentences. Empty for send.",
		},
		"suggested_prompt": map[string]any{
			"type":        "string",
			"description": "clarify: the request rewritten to be clear. Empty for send.",
		},
		"fields": map[string]any{
			"type":        "array",
			"description": "clarify: inputs for the missing information, at most 4. Empty for send.",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"label", "description", "type", "options", "default"},
				"properties": map[string]any{
					"label":       map[string]any{"type": "string"},
					"description": map[string]any{"type": "string"},
					"type": map[string]any{
						"type": "string",
						"enum": []string{"text", "number", "channel", "choice"},
					},
					"options": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "choice: the answers to pick from. Empty otherwise.",
					},
					"default": map[string]any{
						"type":        "string",
						"description": "A suggested value, or empty.",
					},
				},
			},
		},
	},
}

var checkInstructions = checkInstructionsText + "\n\n" + flow.CatalogSummary

const checkInstructionsText = `You help users of Kite, a no-code Discord bot builder, many of them young, ask its flow AI for what they want. A flow is a graph of blocks that starts with a slash command, a Discord event, a schedule, or a click on a button or select menu. You get the user's first message to the flow AI and the flow they are working on. The flow AI then changes the flow, and every change uses up one of the user's limited prompts, so the request should be clear and complete before it goes out.

Reply with verdict "send" if the flow AI can build something useful from the request. This should be the answer for most requests. Don't ask about details with a sensible default or that the flow AI can choose, like texts, colors, or a command name the user doesn't seem to care about, or that it can take from the interaction or from a command argument, like the user who ran the command, the channel it was run in, or the member to ban. Questions and messages that don't ask for a change are "send" too.

Reply with verdict "clarify" only if the request is too vague to build anything useful, or needs something only the user knows, like a specific channel, role or user that isn't the one of the interaction. Then:
- message: one or two short, friendly sentences about what's missing, in the user's language and in simple words.
- suggested_prompt: the request rewritten to be clear and specific, in the user's language. Keep everything they asked for and add nothing they didn't. Refer to the fields instead of making up values, like "in the channel I picked below".
- fields: at most 4 inputs for the missing information. label is short, description helps to fill it in. Use type channel for a channel of the server, choice with options when there are a few sensible answers, and number or text otherwise. For a role or user, use text and explain in the description how to get its ID: turn on Developer Mode in Discord's settings, then right-click it and pick Copy ID. default is a suggested value, or empty.

What the flow AI can build with, as blocks with the settings they need:`
