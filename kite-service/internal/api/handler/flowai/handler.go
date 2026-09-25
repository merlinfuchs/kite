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
	Check(ctx context.Context, req flowai.CheckRequest) (*flowai.CheckResponse, error)
}

// repairWindow is how long after a prompt its edits can be repaired.
const repairWindow = time.Hour

// Answers without edits don't count as prompts, so the AI can ask what's
// missing for free. They are limited to this many times the plan's prompts,
// so the AI can't be used as a free chatbot.
const answerLimitFactor = 3

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
	count, err := h.promptCount(c)
	if err != nil {
		return nil, err
	}

	return &wire.FlowAIUsage{PromptsUsed: count.Edited, PromptsLimit: c.Features.MaxAIPromptsPerMonth}, nil
}

func (h *FlowAIHandler) HandleFlowAIChat(c *handler.Context, req wire.FlowAIChatRequest) (*wire.FlowAIChatResponse, error) {
	if err := h.checkAvailable(c); err != nil {
		return nil, err
	}
	limit := c.Features.MaxAIPromptsPerMonth
	count, err := h.promptCount(c)
	if err != nil {
		return nil, err
	}
	used := count.Edited

	// The prompt is recorded, and counted as edited, before the model is
	// called, so concurrent requests can't all pass the limits.
	now := time.Now().UTC()
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
		// Prompts whose answer made no edits don't count, so their repairs
		// can't be used to get edits for free.
		if !prompt.Edited {
			return nil, handler.ErrBadRequest("nothing_to_repair", "The prompt made no changes to repair.")
		}
		if now.Sub(prompt.CreatedAt) > repairWindow {
			return nil, handler.ErrBadRequest("repair_expired", "The prompt is too old to be repaired.")
		}

		started, err := h.promptStore.StartFlowAIPromptRound(c.Context(), c.App.ID, prompt.ID, 1+h.maxRepairs, now)
		if err != nil {
			return nil, fmt.Errorf("failed to start flow AI prompt round: %w", err)
		}
		if !started {
			return nil, handler.ErrBadRequest("repair_limit", "The AI couldn't fix its changes. Try describing the change differently.")
		}
	} else {
		if used >= limit {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("You've used all %d AI prompts for this month.", limit))
		}
		if count.Total >= answerLimitFactor*limit {
			return nil, handler.ErrBadRequest("resource_limit", "You've asked the AI too many questions this month.")
		}

		prompt = &model.FlowAIPrompt{
			ID:        util.UniqueID(),
			AppID:     c.App.ID,
			UserID:    c.Session.UserID,
			Model:     h.assistant.Model(),
			Prompt:    req.Messages[len(req.Messages)-1].Content,
			Rounds:    1,
			Edited:    true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := h.promptStore.CreateFlowAIPrompt(c.Context(), prompt); err != nil {
			return nil, fmt.Errorf("failed to create flow AI prompt: %w", err)
		}
		used++
	}

	messages := make([]flowai.Message, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = flowai.Message{Role: m.Role, Content: m.Content}
	}

	res, err := h.assistant.Respond(c.Context(), flowai.Request{
		Flow:     req.Flow,
		Messages: messages,
		Issues:   req.Issues,
		AppID:    c.App.ID,
		UserID:   c.Session.UserID,
	})
	// Answers without edits don't count. Ones that can't be used do, as they
	// cost as much, and so do ones whose edits were all invalid, as they are
	// repaired.
	edited := isRepair || err != nil || len(res.Edits) > 0 || len(res.Issues) > 0
	if res != nil {
		// Failing to record the usage shouldn't lose the answer.
		if err := h.promptStore.AddFlowAIPromptUsage(c.Context(), c.App.ID, prompt.ID, res.Usage, edited, time.Now().UTC()); err != nil {
			slog.Error("Failed to add flow AI prompt usage", slog.String("app_id", c.App.ID), slog.Any("error", err))
		}
	}
	if err != nil {
		// Prompts the model didn't answer at all don't count.
		if res == nil && !isRepair {
			if err := h.promptStore.DeleteFlowAIPrompt(c.Context(), c.App.ID, prompt.ID); err != nil {
				slog.Error("Failed to delete flow AI prompt", slog.String("app_id", c.App.ID), slog.Any("error", err))
			}
		}

		var resErr *flowai.ErrResponse
		if errors.As(err, &resErr) {
			return nil, handler.ErrServiceUnavailable("flow_ai_failed", resErr.Message)
		}
		slog.Error("Failed to get flow AI response", slog.String("app_id", c.App.ID), slog.Any("error", err))
		return nil, handler.ErrServiceUnavailable("flow_ai_unavailable", "The flow AI isn't available right now. Please try again later.")
	}

	if !edited {
		used--
	}

	return &wire.FlowAIChatResponse{
		PromptID: prompt.ID,
		Message:  res.Message,
		Edits:    res.Edits,
		Issues:   res.Issues,
		Usage:    wire.FlowAIUsage{PromptsUsed: used, PromptsLimit: limit},
	}, nil
}

// HandleFlowAICheck checks the first prompt of a chat with a cheaper model.
// It doesn't count as a prompt.
func (h *FlowAIHandler) HandleFlowAICheck(c *handler.Context, req wire.FlowAICheckRequest) (*wire.FlowAICheckResponse, error) {
	if err := h.checkAvailable(c); err != nil {
		return nil, err
	}

	res, err := h.assistant.Check(c.Context(), flowai.CheckRequest{
		Flow:   req.Flow,
		Prompt: req.Prompt,
		AppID:  c.App.ID,
		UserID: c.Session.UserID,
	})
	if err != nil {
		slog.Error("Failed to check flow AI prompt", slog.String("app_id", c.App.ID), slog.Any("error", err))
		return nil, handler.ErrServiceUnavailable("flow_ai_unavailable", "The prompt couldn't be checked.")
	}

	fields := make([]wire.FlowAICheckField, len(res.Fields))
	for i, f := range res.Fields {
		fields[i] = wire.FlowAICheckField(f)
	}
	return &wire.FlowAICheckResponse{
		Verdict:         res.Verdict,
		Message:         res.Message,
		SuggestedPrompt: res.SuggestedPrompt,
		Fields:          fields,
	}, nil
}

func (h *FlowAIHandler) checkAvailable(c *handler.Context) error {
	if h.assistant == nil {
		return handler.ErrServiceUnavailable("flow_ai_unavailable", "The flow AI isn't set up on this server.")
	}
	// Unlike other limits, 0 means none, so plans need to opt in.
	if c.Features.MaxAIPromptsPerMonth == 0 {
		return handler.ErrForbidden("feature_unavailable", "Your plan doesn't include the flow AI.")
	}
	return nil
}

func (h *FlowAIHandler) promptCount(c *handler.Context) (model.FlowAIPromptCount, error) {
	start, end := util.StartAndEndOfMonth(time.Now().UTC())
	count, err := h.promptStore.CountFlowAIPromptsBetween(c.Context(), c.App.ID, start, end)
	if err != nil {
		return count, fmt.Errorf("failed to count flow AI prompts: %w", err)
	}
	return count, nil
}
