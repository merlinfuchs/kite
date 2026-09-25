package wire

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type FlowAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (m FlowAIChatMessage) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Role, validation.Required, validation.In("user", "assistant")),
		// The AI can answer with edits alone.
		validation.Field(&m.Content, validation.When(m.Role == "user", validation.Required), validation.Length(0, 4000)),
	)
}

type FlowAIChatRequest struct {
	// Flow is the flow as serialized by the editor.
	Flow     string              `json:"flow"`
	Messages []FlowAIChatMessage `json:"messages"`
	// RepairPromptID asks to fix the issues the editor found with the edits
	// of an earlier prompt. Repairs don't count as new prompts.
	RepairPromptID string   `json:"repair_prompt_id"`
	Issues         []string `json:"issues"`
}

func (req FlowAIChatRequest) Validate() error {
	err := validation.ValidateStruct(&req,
		validation.Field(&req.Flow, validation.Required, validation.Length(1, 100_000)),
		validation.Field(&req.Messages, validation.Required, validation.Length(1, 20)),
		validation.Field(&req.Issues,
			validation.When(req.RepairPromptID != "", validation.Required).Else(validation.Empty),
			validation.Length(0, 50),
			validation.Each(validation.Length(1, 2000)),
		),
	)
	if err != nil {
		return err
	}

	// A repair follows the answer it fixes, anything else asks something new.
	lastRole := req.Messages[len(req.Messages)-1].Role
	if req.RepairPromptID == "" && lastRole != "user" {
		return validation.Errors{"messages": errors.New("the last message must be from the user")}
	}
	if req.RepairPromptID != "" && lastRole != "assistant" {
		return validation.Errors{"messages": errors.New("the last message of a repair must be from the assistant")}
	}
	return nil
}

type FlowAIChatResponse struct {
	PromptID string `json:"prompt_id"`
	// Message is Markdown.
	Message string `json:"message"`
	// BuildPrompt is a request the user can send to make the change the
	// message suggests, if any.
	BuildPrompt string `json:"build_prompt"`
	// Edits are applied with the editor's applyFlowEdits.
	Edits []map[string]any `json:"edits"`
	// Issues are problems with edits that had to be skipped. They are fixed
	// with a repair, like the problems the editor finds.
	Issues []string    `json:"issues"`
	Usage  FlowAIUsage `json:"usage"`
}

type FlowAIUsage struct {
	PromptsUsed  int `json:"prompts_used"`
	PromptsLimit int `json:"prompts_limit"`
}

type FlowAIUsageGetResponse = FlowAIUsage

type FlowAICheckRequest struct {
	// Flow is the start of the flow as serialized by the editor, which is
	// enough to check a prompt.
	Flow   string `json:"flow"`
	Prompt string `json:"prompt"`
}

func (req FlowAICheckRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Flow, validation.Required, validation.Length(1, 10_000)),
		validation.Field(&req.Prompt, validation.Required, validation.Length(1, 4000)),
	)
}

// FlowAICheckResponse says whether a prompt is ready to be sent. If Verdict
// is "clarify", it suggests a clearer prompt and fields for what's missing.
type FlowAICheckResponse struct {
	Verdict         string             `json:"verdict"`
	Message         string             `json:"message"`
	SuggestedPrompt string             `json:"suggested_prompt"`
	Fields          []FlowAICheckField `json:"fields"`
}

type FlowAICheckField struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	// Type is "text", "number", "channel" or "choice".
	Type    string   `json:"type"`
	Options []string `json:"options"`
	Default string   `json:"default"`
}
