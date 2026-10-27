package flow

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionAIChatCompletion:       executeActionAIChatCompletion,
		FlowNodeTypeActionAISearchWeb:            executeActionAIChatCompletion,
		FlowNodeTypeControlConditionAIChoice:     executeControlConditionAIChoice,
		FlowNodeTypeControlConditionItemAIChoice: executeControlConditionItemAIChoice,
		FlowNodeTypeControlAICheck:               executeControlAICheck,
		FlowNodeTypeActionAIScore:                executeActionAIScore,
	})
}

// aiDecisionModel answers AI Choice, Check and Score. It's pinned, so a new
// version can't change what flows decide or cost. It's billed by input token,
// which aiDecisionMaxInputLength bounds, so the blocks only cost 1 credit.
const aiDecisionModel = "typesafe/jev-1.14"

// aiDecisionMaxInputLength bounds the bytes of the question, input, options
// and levels together, more than a Discord message has. It counts bytes rather
// than characters as a token is never more than a byte, and a character can be
// up to 4 tokens.
const aiDecisionMaxInputLength = 8000

const aiScoreMaxLevels = 10

// aiChatMaxInputLength bounds the bytes of the prompt and system prompt
// together, after placeholders are filled in. The blocks cost the same
// credits however long the prompt is, so it bounds what a run can cost Kite.
const aiChatMaxInputLength = 8000

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
		return err
	}

	prompt, err := ctx.EvalTemplate(data.Prompt)
	if err != nil {
		return err
	}

	if length := len(prompt.String()) + len(systemPrompt.String()); length > aiChatMaxInputLength {
		return fmt.Errorf(
			"prompt and system prompt are %d bytes, more than the limit of %d",
			length, aiChatMaxInputLength,
		)
	}

	maxCompletionTokens, err := ctx.EvalTemplate(data.MaxCompletionTokens)
	if err != nil {
		return err
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
		return err
	}

	ctx.StoreNodeResult(n, thing.NewString(response))
	return n.ExecuteChildren(ctx)
}

func aiDecisionData(n *CompiledFlowNode) (*AIDecisionData, error) {
	if n.Data.AIDecisionData == nil {
		return nil, &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "ai_decision_data is nil",
		}
	}
	return n.Data.AIDecisionData, nil
}

// decide evaluates the question and input of a decision block and asks the
// model.
func decide(ctx *FlowContext, data *AIDecisionData, opts provider.DecideOpts) (provider.Decision, error) {
	question, err := ctx.EvalTemplate(data.Question)
	if err != nil {
		return provider.Decision{}, err
	}
	input, err := ctx.EvalTemplate(data.Input)
	if err != nil {
		return provider.Decision{}, err
	}

	opts.Model = aiDecisionModel
	opts.Question = question.String()
	opts.Input = input.String()
	if strings.TrimSpace(opts.Question) == "" {
		return provider.Decision{}, errors.New("question is empty")
	}
	// Like a message without content, which the model would still judge.
	if strings.TrimSpace(opts.Input) == "" {
		return provider.Decision{}, errors.New("input is empty")
	}

	length := len(opts.Question) + len(opts.Input)
	for name, guidance := range opts.Options {
		length += len(name) + len(guidance)
	}
	for _, level := range opts.Levels {
		length += len(level)
	}
	if length > aiDecisionMaxInputLength {
		return provider.Decision{}, fmt.Errorf(
			"question, input and options are %d bytes, more than the limit of %d",
			length, aiDecisionMaxInputLength,
		)
	}

	return ctx.Decision.Decide(ctx, opts)
}

// evalProbability evaluates a threshold from 0 to 1, or returns fallback if
// it's unset.
func evalProbability(ctx *FlowContext, name string, template string, fallback float64) (float64, error) {
	if template == "" {
		return fallback, nil
	}

	value, err := ctx.EvalTemplate(template)
	if err != nil {
		return 0, err
	}

	s := strings.TrimSpace(value.String())
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || f < 0 || f > 1 {
		return 0, fmt.Errorf("%s must be a number from 0 to 1, got %q", name, s)
	}
	return f, nil
}

func executeControlConditionAIChoice(n *CompiledFlowNode, ctx *FlowContext) error {
	data, err := aiDecisionData(n)
	if err != nil {
		return err
	}

	options := make(map[string]string)
	items := make(map[string]*CompiledFlowNode)
	var elseNode *CompiledFlowNode
	for _, child := range n.Children.Default {
		switch child.Type {
		case FlowNodeTypeControlConditionItemElse:
			elseNode = child
		case FlowNodeTypeControlConditionItemAIChoice:
			name := strings.TrimSpace(child.Data.AIChoiceName)
			if name == "" {
				return errors.New("every option needs a name")
			}
			if _, ok := items[name]; ok {
				return fmt.Errorf("option %q exists twice", name)
			}
			options[name] = child.Data.AIChoiceDescription
			items[name] = child
		}
	}
	if len(items) == 0 {
		return errors.New("add at least one option")
	}

	minConfidence, err := evalProbability(ctx, "min confidence", data.MinConfidence, 0)
	if err != nil {
		return err
	}

	decision, err := decide(ctx, data, provider.DecideOpts{
		Type:    provider.DecisionTypeChoice,
		Options: options,
	})
	if err != nil {
		return err
	}

	picked, ok := items[decision.Choice]
	if !ok {
		decision.Choice = ""
	}
	ctx.StoreNodeResult(n, thing.NewString(decision.Choice))

	// The options have nothing to compare, so the picked one runs from here.
	// The else branch checks ConditionItemMet, which is set every time for
	// when the block runs again, like in a loop.
	if decision.Confidence < minConfidence {
		picked = nil
	}
	ctx.GetNodeState(n.ID).ConditionItemMet = picked != nil

	next := picked
	if next == nil {
		next = elseNode
	}
	if next != nil {
		if err := next.ExecuteConditionItem(ctx); err != nil {
			return err
		}
	}
	return nil
}

func executeControlConditionItemAIChoice(n *CompiledFlowNode, ctx *FlowContext) error {
	return n.ExecuteChildren(ctx)
}

// executeControlAICheck continues with the yes, no or unsure branch,
// depending on how likely the model thinks the answer is yes.
func executeControlAICheck(n *CompiledFlowNode, ctx *FlowContext) error {
	data, err := aiDecisionData(n)
	if err != nil {
		return err
	}

	// An unset threshold is the other one, so setting one doesn't add an
	// unsure range.
	yesTemplate, noTemplate := data.YesThreshold, data.NoThreshold
	if yesTemplate == "" {
		yesTemplate = noTemplate
	}
	if noTemplate == "" {
		noTemplate = yesTemplate
	}

	yesThreshold, err := evalProbability(ctx, "yes threshold", yesTemplate, 0.5)
	if err != nil {
		return err
	}
	noThreshold, err := evalProbability(ctx, "no threshold", noTemplate, 0.5)
	if err != nil {
		return err
	}
	if noThreshold > yesThreshold {
		return errors.New("no threshold can't be above yes threshold")
	}

	decision, err := decide(ctx, data, provider.DecideOpts{
		Type:    provider.DecisionTypeNoul,
		Options: map[string]string{"true": "Yes", "false": "No"},
	})
	if err != nil {
		return err
	}

	ctx.StoreNodeResult(n, thing.NewFloat(decision.Noul))

	switch {
	case decision.Noul >= yesThreshold:
		return n.ExecuteChildrenByHandle(ctx, "yes")
	case decision.Noul < noThreshold:
		return n.ExecuteChildrenByHandle(ctx, "no")
	default:
		return n.ExecuteChildrenByHandle(ctx, "unsure")
	}
}

func executeActionAIScore(n *CompiledFlowNode, ctx *FlowContext) error {
	data, err := aiDecisionData(n)
	if err != nil {
		return err
	}

	if len(data.Levels) < 2 || len(data.Levels) > aiScoreMaxLevels {
		return fmt.Errorf("needs 2 to %d levels, has %d", aiScoreMaxLevels, len(data.Levels))
	}
	for _, level := range data.Levels {
		if strings.TrimSpace(level) == "" {
			return errors.New("every level needs a description")
		}
	}

	decision, err := decide(ctx, data, provider.DecideOpts{
		Type:   provider.DecisionTypeScore,
		Levels: data.Levels,
	})
	if err != nil {
		return err
	}

	// The model counts levels from 0, users from 1.
	ctx.StoreNodeResult(n, thing.NewFloat(decision.Score+1))
	return n.ExecuteChildren(ctx)
}
