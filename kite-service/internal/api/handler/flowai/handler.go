package flowai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/flowai"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// Assistant is what the handler needs of flowai.Assistant.
type Assistant interface {
	Model() string
	Respond(ctx context.Context, req flowai.Request) (*flowai.Response, error)
}

type FlowAIHandler struct {
	promptStore store.FlowAIPromptStore
	// assistant is nil if no OpenAI API key is configured.
	assistant  Assistant
	maxRepairs int
}

func NewFlowAIHandler(promptStore store.FlowAIPromptStore, assistant *flowai.Assistant, maxRepairs int) *FlowAIHandler {
	h := &FlowAIHandler{
		promptStore: promptStore,
		maxRepairs:  maxRepairs,
	}
	// A nil pointer in the interface wouldn't compare equal to nil.
	if assistant != nil {
		h.assistant = assistant
	}
	return h
}

func (h *FlowAIHandler) HandleFlowAIUsageGet(c *handler.Context) (*wire.FlowAIUsageGetResponse, error) {
	used, err := h.promptsUsed(c)
	if err != nil {
		return nil, err
	}

	return &wire.FlowAIUsage{PromptsUsed: used, PromptsLimit: c.Features.MaxAIPromptsPerMonth}, nil
}

func (h *FlowAIHandler) HandleFlowAIChat(c *handler.Context, req wire.FlowAIChatRequest) (*wire.FlowAIChatResponse, error) {
	if h.assistant == nil {
		return nil, handler.ErrServiceUnavailable("flow_ai_unavailable", "The flow AI isn't set up on this server.")
	}

	limit := c.Features.MaxAIPromptsPerMonth
	used, err := h.promptsUsed(c)
	if err != nil {
		return nil, err
	}

	isRepair := req.RepairPromptID != ""
	var prompt *model.FlowAIPrompt
	if isRepair {
		prompt, err = h.promptStore.FlowAIPrompt(c.Context(), c.App.ID, req.RepairPromptID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return nil, handler.ErrNotFound("unknown_prompt", "Prompt not found")
			}
			return nil, fmt.Errorf("failed to get flow AI prompt: %w", err)
		}
		if prompt.Rounds > h.maxRepairs {
			return nil, handler.ErrBadRequest("repair_limit", "The AI couldn't fix its changes. Try describing the change differently.")
		}
	} else {
		// Unlike other limits, 0 means none, so plans need to opt in.
		if limit == 0 {
			return nil, handler.ErrForbidden("feature_unavailable", "Your plan doesn't include the flow AI.")
		}
		if used >= limit {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("You've used all %d AI prompts for this month.", limit))
		}
	}

	messages := make([]flowai.Message, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = flowai.Message{Role: m.Role, Content: m.Content}
	}

	res, err := h.assistant.Respond(c.Context(), flowai.Request{
		FlowType: req.FlowType,
		Flow:     req.Flow,
		Messages: messages,
		Issues:   req.Issues,
		AppID:    c.App.ID,
		UserID:   c.Session.UserID,
	})
	if err != nil {
		var resErr *flowai.ErrResponse
		if errors.As(err, &resErr) {
			return nil, handler.ErrServiceUnavailable("flow_ai_failed", resErr.Message)
		}
		slog.Error("Failed to get flow AI response", slog.String("app_id", c.App.ID), slog.Any("error", err))
		return nil, handler.ErrServiceUnavailable("flow_ai_unavailable", "The flow AI isn't available right now. Please try again later.")
	}

	// Only answered prompts count, so failed ones don't use up the limit.
	now := time.Now().UTC()
	if isRepair {
		if err := h.promptStore.AddFlowAIPromptRound(c.Context(), c.App.ID, prompt.ID, res.Usage, now); err != nil {
			return nil, fmt.Errorf("failed to add flow AI prompt round: %w", err)
		}
	} else {
		prompt = &model.FlowAIPrompt{
			ID:        util.UniqueID(),
			AppID:     c.App.ID,
			UserID:    c.Session.UserID,
			Model:     h.assistant.Model(),
			Rounds:    1,
			Usage:     res.Usage,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := h.promptStore.CreateFlowAIPrompt(c.Context(), prompt); err != nil {
			return nil, fmt.Errorf("failed to create flow AI prompt: %w", err)
		}
		used++
	}

	return &wire.FlowAIChatResponse{
		PromptID: prompt.ID,
		Message:  res.Message,
		Edits:    res.Edits,
		Issues:   res.Issues,
		Usage:    wire.FlowAIUsage{PromptsUsed: used, PromptsLimit: limit},
	}, nil
}

func (h *FlowAIHandler) promptsUsed(c *handler.Context) (int, error) {
	start, end := util.StartAndEndOfMonth(time.Now().UTC())
	used, err := h.promptStore.CountFlowAIPromptsBetween(c.Context(), c.App.ID, start, end)
	if err != nil {
		return 0, fmt.Errorf("failed to count flow AI prompts: %w", err)
	}
	return used, nil
}
