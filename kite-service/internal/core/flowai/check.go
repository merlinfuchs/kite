package flowai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/responses"
	"github.com/openai/openai-go/v2/shared"
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
	resp, err := a.client.Responses.New(ctx, responses.ResponseNewParams{
		Model:        a.config.CheckModel,
		Instructions: openai.String(checkInstructions),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(fmt.Sprintf("Current flow:\n%s\n\nThe user's request:\n%s", req.Flow, req.Prompt)),
		},
		MaxOutputTokens: openai.Int(int64(a.config.CheckMaxOutputTokens)),
		Reasoning: shared.ReasoningParam{
			Effort: shared.ReasoningEffort(a.config.CheckReasoningEffort),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "prompt_check",
					Schema: checkSchema,
					Strict: openai.Bool(true),
				},
			},
		},
		PromptCacheKey:   openai.String("kite-flow-ai-check"),
		SafetyIdentifier: openai.String(safetyIdentifier(req.UserID)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create response: %w", err)
	}

	slog.Info(
		"Flow AI check",
		slog.String("app_id", req.AppID),
		slog.String("status", string(resp.Status)),
		slog.Int("input_tokens", int(resp.Usage.InputTokens)),
		slog.Int("cached_input_tokens", int(resp.Usage.InputTokensDetails.CachedTokens)),
		slog.Int("output_tokens", int(resp.Usage.OutputTokens)),
	)
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

var checkInstructions = checkInstructionsText + "\n\n" + catalogSummary()

const checkInstructionsText = `You help users of Kite, a no-code Discord bot builder, many of them young, ask its flow AI for what they want. A flow is a graph of blocks that starts with a slash command, a Discord event, a schedule, or a click on a button or select menu. You get the user's first message to the flow AI and the flow they are working on. The flow AI then changes the flow, and every change uses up one of the user's limited prompts, so the request should be clear and complete before it goes out.

Reply with verdict "send" if the flow AI can build something useful from the request. This should be the answer for most requests. Don't ask about details with a sensible default or that the flow AI can choose, like texts, colors, or a command name the user doesn't seem to care about, or that it can take from the interaction or from a command argument, like the user who ran the command, the channel it was run in, or the member to ban. Questions and messages that don't ask for a change are "send" too.

Reply with verdict "clarify" only if the request is too vague to build anything useful, or needs something only the user knows, like a specific channel, role or user that isn't the one of the interaction. Then:
- message: one or two short, friendly sentences about what's missing, in the user's language and in simple words.
- suggested_prompt: the request rewritten to be clear and specific, in the user's language. Keep everything they asked for and add nothing they didn't. Refer to the fields instead of making up values, like "in the channel I picked below".
- fields: at most 4 inputs for the missing information. label is short, description helps to fill it in. Use type channel for a channel of the server, choice with options when there are a few sensible answers, and number or text otherwise. For a role or user, use text and explain in the description how to get its ID: turn on Developer Mode in Discord's settings, then right-click it and pick Copy ID. default is a suggested value, or empty.

What the flow AI can build with, as blocks with the settings they need:`

// catalogSummary lists the blocks the flow AI can add, shorter than the
// catalog the flow AI gets, so the check stays cheap.
func catalogSummary() string {
	var catalog struct {
		Nodes map[string]struct {
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Contexts    []string `json:"contexts"`
			Fixed       bool     `json:"fixed"`
			DataSchema  struct {
				Required   []string `json:"required"`
				Properties map[string]struct {
					Description string `json:"description"`
				} `json:"properties"`
			} `json:"data_schema"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(flow.CatalogJSON, &catalog); err != nil {
		panic(err)
	}

	types := make([]string, 0, len(catalog.Nodes))
	flowTypes := map[string]bool{}
	for t, node := range catalog.Nodes {
		types = append(types, t)
		for _, c := range node.Contexts {
			flowTypes[c] = true
		}
	}
	sort.Strings(types)

	var b strings.Builder
	for _, t := range types {
		node := catalog.Nodes[t]
		// Entries come with the flow, and fixed blocks with their owner.
		if strings.HasPrefix(t, "entry_") || node.Fixed {
			continue
		}

		fmt.Fprintf(&b, "- %s: %s", node.Title, node.Description)
		if len(node.Contexts) < len(flowTypes) {
			fmt.Fprintf(&b, " (only in %s flows)", strings.Join(node.Contexts, ", "))
		}
		for i, name := range node.DataSchema.Required {
			if i == 0 {
				b.WriteString(". Needs ")
			} else {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%s (%s)", name, strings.TrimSuffix(node.DataSchema.Properties[name].Description, "."))
		}
		b.WriteString("\n")
	}
	return b.String()
}
