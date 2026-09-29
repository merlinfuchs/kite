package wire

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/core/flowai"
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

func (req FlowAIChatRequest) AssistantMessages() []flowai.Message {
	res := make([]flowai.Message, len(req.Messages))
	for i, m := range req.Messages {
		res[i] = flowai.Message{Role: m.Role, Content: m.Content}
	}
	return res
}

type FlowAIChatResponse struct {
	PromptID string `json:"prompt_id"`
	// Message is Markdown.
	Message string `json:"message"`
	// BuildPrompt is a request the user can send to make the change the
	// message suggests, if any.
	BuildPrompt string `json:"build_prompt"`
	// Fields ask the user for what the AI needs but only they know.
	Fields []FlowAIField `json:"fields"`
	// Edits are applied with the editor's applyFlowEdits.
	Edits []map[string]any `json:"edits"`
	Usage FlowAIUsage      `json:"usage"`
}

type FlowAIUsage struct {
	PromptsUsed  int `json:"prompts_used"`
	PromptsLimit int `json:"prompts_limit"`
	// Answers are all prompts, including ones without edits, which don't
	// count as prompts but are limited too.
	AnswersUsed  int `json:"answers_used"`
	AnswersLimit int `json:"answers_limit"`
}

type FlowAIUsageGetResponse = FlowAIUsage

// FlowAIField asks the user for something only they know, like a channel.
type FlowAIField struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	// Type is "text", "number", "channel", "category", "role" or "choice".
	Type    string   `json:"type"`
	Options []string `json:"options"`
	Default string   `json:"default"`
}

func FlowAIFieldsToWire(fields []flowai.Field) []FlowAIField {
	res := make([]FlowAIField, len(fields))
	for i, f := range fields {
		res[i] = FlowAIField(f)
	}
	return res
}
