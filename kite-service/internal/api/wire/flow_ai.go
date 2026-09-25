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
		validation.Field(&m.Content, validation.Required, validation.Length(1, 4000)),
	)
}

type FlowAIChatRequest struct {
	FlowType string `json:"flow_type"`
	// Flow is the flow as serialized by the editor.
	Flow     string              `json:"flow"`
	Messages []FlowAIChatMessage `json:"messages"`
	// RepairPromptID asks to fix the issues the editor found with the edits
	// of an earlier prompt. Repairs don't count as new prompts.
	RepairPromptID string   `json:"repair_prompt_id"`
	Issues         []string `json:"issues"`
}

var flowAIFlowTypes = func() []any {
	types := make([]any, len(flowai.FlowTypes))
	for i, t := range flowai.FlowTypes {
		types[i] = t
	}
	return types
}()

func (req FlowAIChatRequest) Validate() error {
	err := validation.ValidateStruct(&req,
		validation.Field(&req.FlowType, validation.Required, validation.In(flowAIFlowTypes...)),
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
	Message  string `json:"message"`
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
