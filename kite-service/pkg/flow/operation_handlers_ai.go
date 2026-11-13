package flow

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// The ai.* operations of code run like the AI blocks and cost the same. Search
// the Web isn't one, as its searches take longer than a run of code can.

// aiAskOperation is ai.ask, Ask AI for code.
func aiAskOperation(call operationCall) (any, error) {
	var arg struct {
		Prompt       string `json:"prompt"`
		SystemPrompt string `json:"system_prompt"`
		Model        string `json:"model"`
		MaxTokens    *int   `json:"max_tokens"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	if strings.TrimSpace(arg.Prompt) == "" {
		return nil, errors.New("missing parameter prompt")
	}
	if arg.Model == "" {
		arg.Model = AIModelSmall
	}
	// Only the tiers, not the models flows stored before them.
	tier, ok := aiModelTiers[arg.Model]
	if !ok {
		return nil, fmt.Errorf("invalid value for parameter model: must be %s, %s or %s", AIModelSmall, AIModelMedium, AIModelLarge)
	}
	maxOutputTokens := aiDefaultOutputTokens
	if arg.MaxTokens != nil {
		if *arg.MaxTokens < 1 || *arg.MaxTokens > aiMaxOutputTokens {
			return nil, fmt.Errorf("invalid value for parameter max_tokens: must be from 1 to %d", aiMaxOutputTokens)
		}
		maxOutputTokens = *arg.MaxTokens
	}

	ask := aiAsk{
		tier:            tier,
		systemPrompt:    arg.SystemPrompt,
		prompt:          arg.Prompt,
		maxOutputTokens: maxOutputTokens,
	}
	if err := ask.validate(); err != nil {
		return nil, err
	}

	// Like the block: the input now, the output once the answer is there,
	// and the most the output can cost has to fit before either.
	if err := call.chargeCredits(ask.credits()+ask.extraInputCredits(), ask.maxExtraOutputCredits()); err != nil {
		return nil, err
	}
	return ask.run(call.ctx, call.chargeMore)
}

// decideForCode checks opts, charges the call and asks the model.
func decideForCode(call operationCall, opts provider.DecideOpts) (provider.Decision, error) {
	if err := checkDecision(&opts); err != nil {
		return provider.Decision{}, err
	}
	if err := call.charge(); err != nil {
		return provider.Decision{}, err
	}
	return call.ctx.Decision.Decide(call.ctx, opts)
}

// aiCheckOperation is ai.check, AI Check for code, which returns how likely
// the answer to the question is yes, from 0 to 1.
func aiCheckOperation(call operationCall) (any, error) {
	var arg struct {
		Question string `json:"question"`
		Input    string `json:"input"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	decision, err := decideForCode(call, provider.DecideOpts{
		Type:     provider.DecisionTypeNoul,
		Question: arg.Question,
		Input:    arg.Input,
		Options:  map[string]string{"true": "Yes", "false": "No"},
	})
	if err != nil {
		return nil, err
	}
	return decision.Noul, nil
}

// aiChooseOperation is ai.choose, AI Choice for code.
func aiChooseOperation(call operationCall) (any, error) {
	var arg struct {
		Question string            `json:"question"`
		Input    string            `json:"input"`
		Options  map[string]string `json:"options"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	decision, err := decideForCode(call, provider.DecideOpts{
		Type:     provider.DecisionTypeChoice,
		Question: arg.Question,
		Input:    arg.Input,
		Options:  arg.Options,
	})
	if err != nil {
		return nil, err
	}

	res := map[string]any{"choice": nil, "confidence": decision.Confidence}
	if _, ok := arg.Options[decision.Choice]; ok {
		res["choice"] = decision.Choice
	}
	return res, nil
}

// aiScoreOperation is ai.score, AI Score for code, which returns the level
// from 1.
func aiScoreOperation(call operationCall) (any, error) {
	var arg struct {
		Question string   `json:"question"`
		Input    string   `json:"input"`
		Levels   []string `json:"levels"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	decision, err := decideForCode(call, provider.DecideOpts{
		Type:     provider.DecisionTypeScore,
		Question: arg.Question,
		Input:    arg.Input,
		Levels:   arg.Levels,
	})
	if err != nil {
		return nil, err
	}
	// The model counts levels from 0, users from 1.
	return decision.Score + 1, nil
}

// aiTranslateOperation is ai.translate, AI Translate for code.
func aiTranslateOperation(call operationCall) (any, error) {
	var arg struct {
		Text     string `json:"text"`
		Language string `json:"language"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	t, err := newAITranslation(arg.Text, arg.Language)
	if err != nil {
		return nil, err
	}
	if err := call.chargeCredits(t.credits(), 0); err != nil {
		return nil, err
	}
	return t.run(call.ctx)
}
