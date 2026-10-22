package flow

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

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
// which aiDecisionMaxInputLength bounds, so the blocks cost aiDecisionCredits.
const aiDecisionModel = "typesafe/jev-1.14"

// aiDecisionMaxInputLength bounds the characters of the question, input,
// options and levels together, far more than a Discord message has.
const aiDecisionMaxInputLength = 8000

const aiDecisionCredits = 1

const aiScoreMaxLevels = 10

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

	length := utf8.RuneCountInString(opts.Question) + utf8.RuneCountInString(opts.Input)
	for name, guidance := range opts.Options {
		length += utf8.RuneCountInString(name) + utf8.RuneCountInString(guidance)
	}
	for _, level := range opts.Levels {
		length += utf8.RuneCountInString(level)
	}
	if length > aiDecisionMaxInputLength {
		return provider.Decision{}, fmt.Errorf(
			"question, input and options are %d characters, more than the limit of %d",
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
	var elseNode *CompiledFlowNode
	for _, child := range n.Children.Default {
		switch child.Type {
		case FlowNodeTypeControlConditionItemElse:
			elseNode = child
		case FlowNodeTypeControlConditionItemAIChoice:
			name := strings.TrimSpace(child.Data.AIChoiceName)
			if name == "" {
				return traceError(n, errors.New("every option needs a name"))
			}
			if _, ok := options[name]; ok {
				return traceError(n, fmt.Errorf("option %q exists twice", name))
			}
			options[name] = child.Data.AIChoiceDescription
		}
	}
	if len(options) == 0 {
		return traceError(n, errors.New("add at least one option"))
	}

	minConfidence, err := evalProbability(ctx, "min confidence", data.MinConfidence, 0)
	if err != nil {
		return traceError(n, err)
	}

	decision, err := decide(ctx, data, provider.DecideOpts{
		Type:    provider.DecisionTypeChoice,
		Options: options,
	})
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, thing.NewString(decision.Choice))

	// The options compare against the base value, which is empty when the
	// model isn't confident enough, so only the else branch runs. The state
	// is reset for when the block runs again, like in a loop.
	choice := decision.Choice
	if decision.Confidence < minConfidence {
		choice = ""
	}
	ctx.StoreNodeBaseValue(n, thing.NewString(choice))
	ctx.GetNodeState(n.ID).ConditionItemMet = false

	for _, child := range n.Children.Default {
		if child.Type == FlowNodeTypeControlConditionItemElse {
			continue
		}
		if err := child.Execute(ctx); err != nil {
			return traceError(n, err)
		}
	}

	if elseNode != nil {
		if err := elseNode.Execute(ctx); err != nil {
			return traceError(n, err)
		}
	}

	return nil
}

func executeControlConditionItemAIChoice(n *CompiledFlowNode, ctx *FlowContext) error {
	parent := n.FindDirectParentWithType(FlowNodeTypeControlConditionAIChoice)
	if parent == nil {
		return nil
	}

	parentState := ctx.GetNodeState(parent.ID)
	if parentState.ConditionItemMet {
		return nil
	}

	if parentState.ConditionBaseValue.String() != strings.TrimSpace(n.Data.AIChoiceName) {
		return nil
	}

	parentState.ConditionItemMet = true
	return n.ExecuteChildren(ctx)
}

// executeControlAICheck continues with the yes, no or unsure branch,
// depending on how likely the model thinks the answer is yes.
func executeControlAICheck(n *CompiledFlowNode, ctx *FlowContext) error {
	data, err := aiDecisionData(n)
	if err != nil {
		return err
	}

	yesThreshold, err := evalProbability(ctx, "yes threshold", data.YesThreshold, 0.5)
	if err != nil {
		return traceError(n, err)
	}
	noThreshold, err := evalProbability(ctx, "no threshold", data.NoThreshold, 0.5)
	if err != nil {
		return traceError(n, err)
	}
	if noThreshold > yesThreshold {
		return traceError(n, errors.New("no threshold can't be above yes threshold"))
	}

	decision, err := decide(ctx, data, provider.DecideOpts{
		Type:    provider.DecisionTypeNoul,
		Options: map[string]string{"true": "Yes", "false": "No"},
	})
	if err != nil {
		return traceError(n, err)
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
		return traceError(n, fmt.Errorf("needs 2 to %d levels, has %d", aiScoreMaxLevels, len(data.Levels)))
	}
	for _, level := range data.Levels {
		if strings.TrimSpace(level) == "" {
			return traceError(n, errors.New("every level needs a description"))
		}
	}

	decision, err := decide(ctx, data, provider.DecideOpts{
		Type:   provider.DecisionTypeScore,
		Levels: data.Levels,
	})
	if err != nil {
		return traceError(n, err)
	}

	// The model counts levels from 0, users from 1.
	ctx.StoreNodeResult(n, thing.NewFloat(decision.Score+1))
	return n.ExecuteChildren(ctx)
}
