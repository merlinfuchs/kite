package flow

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionAIChatCompletion: executeActionAIChatCompletion,
		FlowNodeTypeActionAISearchWeb:      executeActionAIChatCompletion,
	})
}

func executeActionAIChatCompletion(n *CompiledFlowNode, ctx *FlowContext) error {
	webSearch := n.Type == FlowNodeTypeActionAISearchWeb

	data := n.Data.AIChatCompletionData
	if data == nil || data.Prompt == "" {
		name := "ai_chat_completion_data"
		if webSearch {
			name = "ai_search_web_data"
		}

		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: fmt.Sprintf("%s is nil", name),
		}
	}

	// Checked here as well as at save time: message flows are not validated
	// by the API at all, and flows stored before the allowlist existed have
	// never been through it.
	tier, ok := resolveAIModel(data.Model)
	if !ok {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: fmt.Sprintf("unsupported ai model: %s", data.Model),
		}
	}

	systemPrompt, err := ctx.EvalTemplate(data.SystemPrompt)
	if err != nil {
		return traceError(n, err)
	}

	prompt, err := ctx.EvalTemplate(data.Prompt)
	if err != nil {
		return traceError(n, err)
	}

	maxCompletionTokens, err := ctx.EvalTemplate(data.MaxCompletionTokens)
	if err != nil {
		return traceError(n, err)
	}

	maxOutputTokens := aiMaxOutputTokens
	if limit := int(maxCompletionTokens.Int()); limit > 0 && limit < maxOutputTokens {
		maxOutputTokens = limit
	}

	opts := provider.CreateResponseOpts{
		Model:           tier.Model,
		ReasoningEffort: tier.ReasoningEffort,
		Prompt:          prompt.String(),
		SystemPrompt:    systemPrompt.String(),
		MaxOutputTokens: maxOutputTokens,
	}
	if webSearch {
		opts.Tools = []provider.AIToolType{provider.AIToolTypeWebSearch}
		opts.MaxToolCalls = aiMaxWebSearches
	}

	response, err := ctx.AI.CreateResponse(ctx, opts)
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, thing.NewString(response))
	return n.ExecuteChildren(ctx)
}
