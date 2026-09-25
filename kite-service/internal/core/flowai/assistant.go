package flowai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/responses"
	"github.com/openai/openai-go/v2/shared"
)

// maxHistory is how many earlier chat messages are sent along, so long chats
// don't grow the cost of every prompt.
const maxHistory = 10

type Config struct {
	Model           string
	ReasoningEffort string
	MaxOutputTokens int
}

// Assistant asks the model for edits to a flow.
type Assistant struct {
	client *openai.Client
	config Config
}

func NewAssistant(client *openai.Client, config Config) *Assistant {
	return &Assistant{client: client, config: config}
}

func (a *Assistant) Model() string {
	return a.config.Model
}

type Message struct {
	// Role is "user" or "assistant".
	Role    string
	Content string
}

type Request struct {
	FlowType string
	// Flow is the flow as serialized by the editor.
	Flow string
	// Messages is the chat so far, oldest first. The last one is the user's
	// current message, unless this is a repair.
	Messages []Message
	// Issues are the problems the editor found with the edits of the last
	// response, which this response should fix.
	Issues []string
	UserID string
}

type Response struct {
	Message string
	// Edits are passed to the editor's applyFlowEdits as they are.
	Edits []map[string]any
	// Issues are problems with edits the model got wrong in a way the editor
	// can't see, e.g. settings that aren't JSON. They are fixed like the
	// editor's own issues.
	Issues []string
	Usage  model.FlowAIUsage
}

// ErrResponse is an error with a message that can be shown to the user.
type ErrResponse struct {
	Message string
}

func (e *ErrResponse) Error() string {
	return e.Message
}

func (a *Assistant) Respond(ctx context.Context, req Request) (*Response, error) {
	resp, err := a.client.Responses.New(ctx, a.params(req))
	if err != nil {
		return nil, fmt.Errorf("failed to create response: %w", err)
	}

	usage := model.FlowAIUsage{
		InputTokens:       int(resp.Usage.InputTokens),
		CachedInputTokens: int(resp.Usage.InputTokensDetails.CachedTokens),
		OutputTokens:      int(resp.Usage.OutputTokens),
	}
	if resp.Status != responses.ResponseStatusCompleted {
		return &Response{Usage: usage}, &ErrResponse{
			Message: "The AI's answer was cut off. Try asking for a smaller change.",
		}
	}

	res, err := parseOutput(resp.OutputText())
	if err != nil {
		return &Response{Usage: usage}, err
	}
	res.Usage = usage
	return res, nil
}

func (a *Assistant) params(req Request) responses.ResponseNewParams {
	messages := req.Messages
	var current string
	if len(req.Issues) > 0 {
		current = "The editor applied your edits and found these problems:\n- " +
			strings.Join(req.Issues, "\n- ") +
			"\n\nFix them with further edits."
	} else if len(messages) > 0 {
		current = messages[len(messages)-1].Content
		messages = messages[:len(messages)-1]
	}
	if len(messages) > maxHistory {
		messages = messages[len(messages)-maxHistory:]
	}

	input := make(responses.ResponseInputParam, 0, len(messages)+1)
	for _, m := range messages {
		role := responses.EasyInputMessageRoleUser
		if m.Role == "assistant" {
			role = responses.EasyInputMessageRoleAssistant
		}
		input = append(input, easyMessage(role, m.Content))
	}
	input = append(input, easyMessage(
		responses.EasyInputMessageRoleUser,
		fmt.Sprintf("Flow type: %s\n\nCurrent flow:\n%s\n\n%s", req.FlowType, req.Flow, current),
	))

	// Identifies users to OpenAI for abuse detection without revealing them.
	userHash := sha256.Sum256([]byte(req.UserID))

	return responses.ResponseNewParams{
		Model:           a.config.Model,
		Instructions:    openai.String(instructions),
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: input},
		MaxOutputTokens: openai.Int(int64(a.config.MaxOutputTokens)),
		Reasoning: shared.ReasoningParam{
			Effort: shared.ReasoningEffort(a.config.ReasoningEffort),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "flow_edits",
					Schema: outputSchema,
					Strict: openai.Bool(true),
				},
			},
		},
		PromptCacheKey:   openai.String("kite-flow-ai"),
		SafetyIdentifier: openai.String(hex.EncodeToString(userHash[:])),
	}
}

func easyMessage(role responses.EasyInputMessageRole, content string) responses.ResponseInputItemUnionParam {
	return responses.ResponseInputItemUnionParam{
		OfMessage: &responses.EasyInputMessageParam{
			Role: role,
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: openai.String(content),
			},
		},
	}
}

// output is the model's answer. Strict structured outputs can't hold objects
// of any shape, so settings come as JSON strings.
type output struct {
	Message string       `json:"message"`
	Edits   []outputEdit `json:"edits"`
}

type outputEdit struct {
	Op        string  `json:"op"`
	Ref       *string `json:"ref"`
	Type      *string `json:"type"`
	ID        *string `json:"id"`
	After     *string `json:"after"`
	Before    *string `json:"before"`
	Handle    *string `json:"handle"`
	Source    *string `json:"source"`
	Target    *string `json:"target"`
	DataJSON  *string `json:"data_json"`
	ItemsJSON *string `json:"items_json"`
	Reconnect *bool   `json:"reconnect"`
}

func parseOutput(text string) (*Response, error) {
	var out output
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return nil, errors.Join(&ErrResponse{Message: "The AI's answer couldn't be read. Please try again."}, err)
	}

	res := &Response{Message: out.Message, Edits: make([]map[string]any, 0, len(out.Edits))}
	for i, e := range out.Edits {
		edit, err := e.toEdit()
		if err != nil {
			res.Issues = append(res.Issues, fmt.Sprintf("Edit %d (%s) was skipped: %s", i+1, e.Op, err))
			continue
		}
		res.Edits = append(res.Edits, edit)
	}
	return res, nil
}

// toEdit converts the edit into the form the editor applies, leaving out the
// fields the model set to null.
func (e outputEdit) toEdit() (map[string]any, error) {
	edit := map[string]any{"op": e.Op}
	for key, value := range map[string]*string{
		"ref":    e.Ref,
		"type":   e.Type,
		"id":     e.ID,
		"after":  e.After,
		"before": e.Before,
		"handle": e.Handle,
		"source": e.Source,
		"target": e.Target,
	} {
		if value != nil {
			edit[key] = *value
		}
	}
	if e.Reconnect != nil {
		edit["reconnect"] = *e.Reconnect
	}

	if e.DataJSON != nil {
		var data map[string]any
		if err := json.Unmarshal([]byte(*e.DataJSON), &data); err != nil || data == nil {
			return nil, fmt.Errorf("data_json isn't a JSON object")
		}
		edit["data"] = data
	}
	if e.ItemsJSON != nil {
		var items []map[string]any
		if err := json.Unmarshal([]byte(*e.ItemsJSON), &items); err != nil {
			return nil, fmt.Errorf("items_json isn't a JSON array of objects")
		}
		edit["items"] = items
	}

	return edit, nil
}
