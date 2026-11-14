package flow

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runJavaScriptAI(t *testing.T, code string, ai provider.AIProvider, decision *fakeDecisionProvider, maxCredits int) (*FlowContext, error) {
	if ai == nil {
		ai = &recordingAIProvider{}
	}
	if decision == nil {
		decision = &fakeDecisionProvider{}
	}
	c := javaScriptTestContext(t, FlowProviders{
		Log:      &provider.MockLogProvider{},
		AI:       ai,
		Decision: decision,
	}, maxCredits, eval.Env{})
	node := &CompiledFlowNode{ID: "js_1", Type: FlowNodeTypeActionJavaScriptRun, Data: FlowNodeData{JavaScriptData: &JavaScriptData{Code: code}}}
	return c, node.Execute(c)
}

func TestJavaScriptRunAIAsk(t *testing.T) {
	ai := &recordingAIProvider{text: "Hi there", outputTokens: 600}
	c, err := runJavaScriptAI(t, `return await ai.ask({prompt: "Say hi", system_prompt: "Be short", model: "medium", max_tokens: 800})`, ai, nil, 1000)
	require.NoError(t, err)
	assert.Equal(t, "Hi there", c.GetNodeResult("js_1").String())
	tier := aiModelTiers[AIModelMedium]
	assert.Equal(t, tier.Model, ai.opts.Model)
	assert.Equal(t, tier.ReasoningEffort, ai.opts.ReasoningEffort)
	assert.Equal(t, "Say hi", ai.opts.Prompt)
	assert.Equal(t, "Be short", ai.opts.SystemPrompt)
	assert.Equal(t, 800, ai.opts.MaxOutputTokens)
	assert.Empty(t, ai.opts.Tools)
	// The block, the medium model, and a second step of answer.
	assert.Equal(t, 10+20+20, c.CreditsUsed())

	ai = &recordingAIProvider{}
	c, err = runJavaScriptAI(t, `return await ai.ask({prompt: "Say hi"})`, ai, nil, 1000)
	require.NoError(t, err)
	assert.Equal(t, aiModelTiers[AIModelSmall].Model, ai.opts.Model)
	assert.Equal(t, aiDefaultOutputTokens, ai.opts.MaxOutputTokens)
	assert.Equal(t, 10+5, c.CreditsUsed())

	// Long prompts cost a step more, like the block.
	c, err = runJavaScriptAI(t, `return await ai.ask({prompt: "x".repeat(9000)})`, &recordingAIProvider{}, nil, 1000)
	require.NoError(t, err)
	assert.Equal(t, 10+5+5, c.CreditsUsed())
}

func TestJavaScriptRunAIErrors(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"no prompt", `await ai.ask({})`, "ai.ask: missing parameter prompt"},
		{"old model", `await ai.ask({prompt: "hi", model: "gpt-4.1"})`, "ai.ask: invalid value for parameter model: must be small, medium or large"},
		{"too many tokens", `await ai.ask({prompt: "hi", max_tokens: 5000})`, "ai.ask: invalid value for parameter max_tokens: must be from 1 to 1000"},
		{"long prompt", `await ai.ask({prompt: "x".repeat(9000), model: "large"})`, "ai.ask: prompt and system prompt are 9000 bytes, more than the limit of 8000 for this model"},
		{"unknown parameter", `await ai.check({question: "Spam?", input: "hi", foo: 1})`, "ai.check: it has no parameter foo"},
		{"empty input", `await ai.check({question: "Spam?", input: " "})`, "ai.check: input is empty"},
		{"no options", `await ai.choose({question: "Which?", input: "hi"})`, "ai.choose: add at least one option"},
		{"one level", `await ai.score({question: "How?", input: "hi", levels: ["low"]})`, "ai.score: needs 2 to 10 levels, has 1"},
		{"language", `await ai.translate({text: "hi", language: "xx"})`, `ai.translate: unsupported language "xx"`},
		{"search", `await ai.searchWeb({prompt: "hi"})`, "ai.searchWeb isn't available"},
		// The most the answer can cost has to fit before anything is paid.
		{"credits", `await ai.ask({prompt: "hi", model: "large", max_tokens: 1000})`, "ai.ask: credits limit exceeded: 150"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := &recordingAIProvider{}
			decision := &fakeDecisionProvider{}
			c, err := runJavaScriptAI(t, tt.code, ai, decision, 150)
			assert.ErrorContains(t, err, tt.want)
			assert.Empty(t, ai.opts.Model)
			assert.Nil(t, decision.opts)
			// Only the block.
			assert.Equal(t, 10, c.CreditsUsed())
		})
	}
}

func TestJavaScriptRunAIDecisions(t *testing.T) {
	decision := &fakeDecisionProvider{decision: provider.Decision{Noul: 0.8}}
	c, err := runJavaScriptAI(t, `return await ai.check({question: "Is it spam?", input: "buy now"})`, nil, decision, 1000)
	require.NoError(t, err)
	assert.Equal(t, 0.8, c.GetNodeResult("js_1").Float())
	assert.Equal(t, provider.DecisionTypeNoul, decision.opts.Type)
	assert.Equal(t, aiDecisionModel, decision.opts.Model)
	assert.Equal(t, "buy now", decision.opts.Input)
	assert.Equal(t, 11, c.CreditsUsed())

	decision = &fakeDecisionProvider{decision: provider.Decision{Choice: "bug", Confidence: 0.9}}
	c, err = runJavaScriptAI(t, `return await ai.choose({question: "What is it?", input: "it crashed", options: {bug: "Something broke", idea: ""}})`, nil, decision, 1000)
	require.NoError(t, err)
	assert.JSONEq(t, `{"choice":"bug","confidence":0.9}`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))
	assert.Equal(t, map[string]string{"bug": "Something broke", "idea": ""}, decision.opts.Options)

	// A choice that isn't an option is none.
	decision = &fakeDecisionProvider{decision: provider.Decision{Choice: "other", Confidence: 0.4}}
	c, err = runJavaScriptAI(t, `return await ai.choose({question: "What is it?", input: "hi", options: {bug: ""}})`, nil, decision, 1000)
	require.NoError(t, err)
	assert.JSONEq(t, `{"choice":null,"confidence":0.4}`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))

	decision = &fakeDecisionProvider{decision: provider.Decision{Score: 2}}
	c, err = runJavaScriptAI(t, `return await ai.score({question: "How rude?", input: "hi", levels: ["nice", "okay", "rude"]})`, nil, decision, 1000)
	require.NoError(t, err)
	assert.Equal(t, 3.0, c.GetNodeResult("js_1").Float())
	assert.Equal(t, []string{"nice", "okay", "rude"}, decision.opts.Levels)
}

func TestJavaScriptRunAITranslate(t *testing.T) {
	ai := &recordingAIProvider{text: `{"source_language":"en","text":"Hola"}`}
	c, err := runJavaScriptAI(t, `return await ai.translate({text: "Hello", language: "es"})`, ai, nil, 1000)
	require.NoError(t, err)
	assert.JSONEq(t, `{"source_language":"en","text":"Hola"}`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))
	assert.True(t, strings.Contains(ai.opts.SystemPrompt, "Spanish"))
	assert.Equal(t, 11, c.CreditsUsed())

	// 1 credit for each 1,000 characters started.
	c, err = runJavaScriptAI(t, `return await ai.translate({text: "x".repeat(2500), language: "es"})`, &recordingAIProvider{text: `{"source_language":"en","text":"x"}`}, nil, 1000)
	require.NoError(t, err)
	assert.Equal(t, 13, c.CreditsUsed())
}

// slowAIProvider answers after a while, so calls of code overlap.
type slowAIProvider struct{}

func (slowAIProvider) CreateResponse(ctx context.Context, opts provider.CreateResponseOpts) (provider.AIResponse, error) {
	time.Sleep(50 * time.Millisecond)
	return provider.AIResponse{Text: "ok"}, nil
}

// Calls at the same time can't use the credits another call holds for the
// longest answer it can get.
func TestJavaScriptRunAIReservesAnswers(t *testing.T) {
	// The block, then 100 for large and up to 100 for a long answer each.
	c, err := runJavaScriptAI(t, `
const results = await Promise.allSettled([1, 2].map(() => ai.ask({prompt: "hi", model: "large", max_tokens: 1000})))
return results.map((r) => r.status === "fulfilled" ? "ok" : r.reason.message)`, slowAIProvider{}, nil, 330)
	require.NoError(t, err)
	assert.ElementsMatch(t, []any{"ok", "ai.ask: credits limit exceeded: 330"}, c.GetNodeResult("js_1").JSONValue())
	// The answer was short, so its reserve wasn't used.
	assert.Equal(t, 110, c.CreditsUsed())
}

func TestJavaScriptRunAIInTestRuns(t *testing.T) {
	c := javaScriptTestContext(t, FlowProviders{Log: &provider.MockLogProvider{}, AI: &recordingAIProvider{}, Decision: &fakeDecisionProvider{}}, 1000, eval.Env{})
	c.Test = NewTestRun(nil)
	node := &CompiledFlowNode{ID: "js_1", Type: FlowNodeTypeActionJavaScriptRun, Data: FlowNodeData{JavaScriptData: &JavaScriptData{Code: `
const errors = []
for (const call of [
  () => ai.ask({prompt: "hi"}),
  () => ai.check({question: "Spam?", input: "hi"}),
  () => ai.translate({text: "hi", language: "es"}),
]) {
  try { await call() } catch (e) { errors.push(e.message) }
}
return errors`}}}
	require.NoError(t, node.Execute(c))
	assert.JSONEq(t, `["ai.ask: test runs don't ask AI","ai.check: test runs don't ask AI","ai.translate: test runs don't ask AI"]`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))
	assert.Equal(t, 10, c.CreditsUsed())
}

// Option names are trimmed like those of the block.
func TestJavaScriptRunAIChooseTrimsOptions(t *testing.T) {
	decision := &fakeDecisionProvider{decision: provider.Decision{Choice: "bug", Confidence: 0.9}}
	c, err := runJavaScriptAI(t, `return await ai.choose({question: "What?", input: "it crashed", options: {"bug ": ""}})`, nil, decision, 1000)
	require.NoError(t, err)
	assert.JSONEq(t, `{"choice":"bug","confidence":0.9}`, toJSON(t, c.GetNodeResult("js_1").JSONValue()))

	_, err = runJavaScriptAI(t, `return await ai.choose({question: "What?", input: "hi", options: {"bug": "", " bug": ""}})`, nil, nil, 1000)
	assert.ErrorContains(t, err, `ai.choose: option "bug" exists twice`)
}
