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
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// Assistant is what the handler needs of flowai.Assistant.
type Assistant interface {
	Model() string
	Respond(ctx context.Context, req flowai.Request) (*flowai.Response, error)
}

// repairWindow is how long after a prompt its edits can be repaired.
const repairWindow = time.Hour

// Answers without edits don't count as prompts, so the AI can ask what's
// missing for free. They are limited to this many times the plan's prompts,
// so the AI can't be used as a free chatbot.
const answerLimitFactor = 3

// VariableStore is what the handler needs of store.VariableStore.
type VariableStore interface {
	VariablesByAppWithoutTotals(ctx context.Context, appID string) ([]*model.Variable, error)
}

// IntegrationStore is what the handler needs to tell which integrations the
// app enabled.
type IntegrationStore interface {
	AppIntegrationCredentials(ctx context.Context, appID string) ([]*model.AppSecret, error)
}

type IntegrationChoiceStore interface {
	AppIntegrations(ctx context.Context, appID string) ([]*model.AppIntegration, error)
}

type FlowAIHandler struct {
	promptStore      store.AssistantPromptStore
	variableStore    VariableStore
	integrationStore IntegrationStore
	choiceStore      IntegrationChoiceStore
	// assistant is nil if no OpenAI API key is configured.
	assistant  Assistant
	maxRepairs int
}

func NewFlowAIHandler(promptStore store.AssistantPromptStore, variableStore VariableStore, integrationStore IntegrationStore, choiceStore IntegrationChoiceStore, assistant *flowai.Assistant, maxRepairs int) *FlowAIHandler {
	h := &FlowAIHandler{
		promptStore:      promptStore,
		variableStore:    variableStore,
		integrationStore: integrationStore,
		choiceStore:      choiceStore,
		maxRepairs:       maxRepairs,
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

	res := usage(count, c.Features.MaxAIPromptsPerMonth)
	return &res, nil
}

func (h *FlowAIHandler) HandleFlowAIChat(c *handler.Context, req wire.FlowAIChatRequest) (*wire.FlowAIChatResponse, error) {
	if h.assistant == nil {
		return nil, handler.ErrServiceUnavailable("flow_ai_unavailable", "The flow AI isn't set up on this server.")
	}
	// Unlike other limits, 0 means none, so plans need to opt in.
	limit := c.Features.MaxAIPromptsPerMonth
	if limit == 0 {
		return nil, handler.ErrForbidden("feature_unavailable", "Your plan doesn't include the flow AI.")
	}

	now := time.Now().UTC()
	isRepair := req.RepairPromptID != ""
	var prompt *model.AssistantPrompt
	if isRepair {
		var err error
		prompt, err = h.promptStore.AssistantPrompt(c.Context(), c.App.ID, req.RepairPromptID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return nil, handler.ErrNotFound("unknown_prompt", "Prompt not found")
			}
			return nil, fmt.Errorf("failed to get assistant prompt: %w", err)
		}
		// Prompts whose answer made no edits don't count, so their repairs
		// can't be used to get edits for free.
		if !prompt.Edited {
			return nil, handler.ErrBadRequest("nothing_to_repair", "The prompt made no changes to repair.")
		}
		if now.Sub(prompt.CreatedAt) > repairWindow {
			return nil, handler.ErrBadRequest("repair_expired", "The prompt is too old to be repaired.")
		}
	}

	// Loaded before the prompt or round is recorded, so failing doesn't use
	// it up.
	variables, err := h.variableStore.VariablesByAppWithoutTotals(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variables: %w", err)
	}
	integrations, err := h.enabledIntegrations(c.Context(), c.App.ID)
	if err != nil {
		return nil, err
	}

	// Recorded even if the client disconnects, so it's counted or given back.
	ctx := context.WithoutCancel(c.Context())
	// Gives back a prompt or repair the model didn't answer.
	giveBack := func() {
		var err error
		if isRepair {
			err = h.promptStore.UndoAssistantPromptRound(ctx, c.App.ID, prompt.ID)
		} else {
			err = h.promptStore.DeleteAssistantPrompt(ctx, c.App.ID, prompt.ID)
		}
		if err != nil {
			slog.Error("Failed to give back assistant prompt", slog.String("app_id", c.App.ID), slog.Any("error", err))
		}
	}

	// The prompt is recorded, and counted as edited, before the limits are
	// checked and the model is called, so concurrent requests can't all pass
	// the limits.
	if isRepair {
		started, err := h.promptStore.StartAssistantPromptRound(c.Context(), c.App.ID, prompt.ID, 1+h.maxRepairs, now)
		if err != nil {
			return nil, fmt.Errorf("failed to start assistant prompt round: %w", err)
		}
		if !started {
			return nil, handler.ErrBadRequest("repair_limit", "The AI couldn't fix its changes. Try describing the change differently.")
		}
	} else {
		prompt = &model.AssistantPrompt{
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
		if err := h.promptStore.CreateAssistantPrompt(c.Context(), prompt); err != nil {
			return nil, fmt.Errorf("failed to create assistant prompt: %w", err)
		}
	}

	count, err := h.promptCount(c)
	if err != nil {
		giveBack()
		return nil, err
	}
	if !isRepair {
		// Concurrent prompts see each other, so near the limit they may all
		// be refused rather than all pass.
		before := model.AssistantPromptCount{Edited: count.Edited - 1, Total: count.Total - 1}
		if err := checkLimits(limit, before); err != nil {
			giveBack()
			return nil, err
		}
	}

	res, err := h.assistant.Respond(c.Context(), flowai.Request{
		Flow:         req.Flow,
		Messages:     req.AssistantMessages(),
		Issues:       req.Issues,
		Variables:    variables,
		Integrations: integrations,
		AppID:        c.App.ID,
		UserID:       c.Session.UserID,
	})
	// Answers without edits don't count. Ones that can't be used do, as they
	// cost as much.
	edited := isRepair || err != nil || len(res.Edits) > 0
	if res != nil {
		// Failing to record the usage shouldn't lose the answer.
		if err := h.promptStore.AddAssistantPromptUsage(ctx, c.App.ID, prompt.ID, res.Usage, edited, time.Now().UTC()); err != nil {
			slog.Error("Failed to add assistant prompt usage", slog.String("app_id", c.App.ID), slog.Any("error", err))
		}
	}
	if err != nil {
		if res == nil {
			giveBack()
		}

		var resErr *flowai.ErrResponse
		if errors.As(err, &resErr) {
			return nil, handler.ErrServiceUnavailable("flow_ai_failed", resErr.Message)
		}
		slog.Error("Failed to get flow AI response", slog.String("app_id", c.App.ID), slog.Any("error", err))
		return nil, handler.ErrServiceUnavailable("flow_ai_unavailable", "The flow AI isn't available right now. Please try again later.")
	}

	if !edited {
		count.Edited--
	}

	return &wire.FlowAIChatResponse{
		PromptID:    prompt.ID,
		Message:     res.Message,
		BuildPrompt: res.BuildPrompt,
		Fields:      wire.FlowAIFieldsToWire(res.Fields),
		Edits:       res.Edits,
		Usage:       usage(count, limit),
	}, nil
}

// checkLimits returns an error if the app can't send another prompt this
// month.
func checkLimits(limit int, count model.AssistantPromptCount) error {
	if count.Edited >= limit {
		return handler.ErrBadRequest("resource_limit", fmt.Sprintf("You've used all %d AI prompts for this month.", limit))
	}
	if count.Total >= answerLimitFactor*limit {
		return handler.ErrBadRequest("resource_limit", "You've asked the AI too many questions this month.")
	}
	return nil
}

func (h *FlowAIHandler) promptCount(c *handler.Context) (model.AssistantPromptCount, error) {
	start, end := util.StartAndEndOfMonth(time.Now().UTC())
	count, err := h.promptStore.CountAssistantPromptsBetween(c.Context(), c.App.ID, start, end)
	if err != nil {
		return count, fmt.Errorf("failed to count assistant prompts: %w", err)
	}
	return count, nil
}

func usage(count model.AssistantPromptCount, limit int) wire.FlowAIUsage {
	return wire.FlowAIUsage{
		PromptsUsed:  count.Edited,
		PromptsLimit: limit,
		AnswersUsed:  count.Total,
		AnswersLimit: answerLimitFactor * limit,
	}
}

func (h *FlowAIHandler) enabledIntegrations(ctx context.Context, appID string) ([]string, error) {
	credentials, err := h.integrationStore.AppIntegrationCredentials(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}
	connected := make([]string, len(credentials))
	for i, credential := range credentials {
		connected[i] = credential.IntegrationID
	}

	rows, err := h.choiceStore.AppIntegrations(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}
	choices := make(map[string]bool, len(rows))
	for _, row := range rows {
		choices[row.IntegrationID] = row.Enabled
	}

	return flow.EnabledIntegrations(connected, choices), nil
}
