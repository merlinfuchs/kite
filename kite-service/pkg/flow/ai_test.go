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

type recordingAIProvider struct {
	opts provider.CreateResponseOpts
}

func (p *recordingAIProvider) CreateResponse(ctx context.Context, opts provider.CreateResponseOpts) (string, error) {
	p.opts = opts
	return "ok", nil
}

func executeAINode(t *testing.T, nodeType FlowNodeType, data AIChatCompletionData) (provider.CreateResponseOpts, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ai := &recordingAIProvider{}
	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{AI: ai, Log: &provider.MockLogProvider{}},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{}),
		nil,
	)
	defer c.Cancel()

	node := &CompiledFlowNode{
		ID:   "0",
		Type: nodeType,
		Data: FlowNodeData{AIChatCompletionData: &data},
	}
	err := node.Execute(c)
	return ai.opts, err
}

func TestAINodeRunsResolvedTier(t *testing.T) {
	cases := []struct {
		model     string
		maxTokens string
		want      aiModelTier
		wantMax   int
	}{
		{"", "", aiModelTiers[AIModelSmall], aiMaxOutputTokens},
		{"gpt-4.1", "100", aiModelTiers[AIModelLarge], 100},
		{AIModelMedium, "5000", aiModelTiers[AIModelMedium], aiMaxOutputTokens},
	}

	for _, tc := range cases {
		opts, err := executeAINode(t, FlowNodeTypeActionAIChatCompletion, AIChatCompletionData{
			Model:               tc.model,
			Prompt:              "hi",
			MaxCompletionTokens: tc.maxTokens,
		})
		require.NoError(t, err)
		assert.Equal(t, tc.want.Model, opts.Model, tc.model)
		assert.Equal(t, tc.want.ReasoningEffort, opts.ReasoningEffort, tc.model)
		assert.Equal(t, tc.wantMax, opts.MaxOutputTokens, tc.model)
		assert.Empty(t, opts.Tools, tc.model)
	}
}

func TestAIWebSearchNodeCapsSearches(t *testing.T) {
	opts, err := executeAINode(t, FlowNodeTypeActionAISearchWeb, AIChatCompletionData{Prompt: "hi"})
	require.NoError(t, err)
	assert.Equal(t, []provider.AIToolType{provider.AIToolTypeWebSearch}, opts.Tools)
	assert.Equal(t, aiMaxWebSearches, opts.MaxToolCalls)
}

func TestAINodeCapsInput(t *testing.T) {
	for _, nodeType := range []FlowNodeType{FlowNodeTypeActionAIChatCompletion, FlowNodeTypeActionAISearchWeb} {
		// Half the limit in characters, but é is 2 bytes.
		opts, err := executeAINode(t, nodeType, AIChatCompletionData{
			Prompt:       "hi",
			SystemPrompt: strings.Repeat("é", aiChatMaxInputLength/2),
		})
		assert.ErrorContains(t, err, "more than the limit of 8000", nodeType)
		assert.Empty(t, opts.Model, nodeType)
	}
}

type fakeDecisionProvider struct {
	decision provider.Decision
	opts     *provider.DecideOpts
}

func (p *fakeDecisionProvider) Decide(ctx context.Context, opts provider.DecideOpts) (provider.Decision, error) {
	p.opts = &opts
	return p.decision, nil
}

// executeDecisionNode runs node, whose branches log their names, with the
// decision the model makes.
func executeDecisionNode(t *testing.T, node *CompiledFlowNode, decision provider.Decision) (*FlowContext, *fakeDecisionProvider, []string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	decisions := &fakeDecisionProvider{decision: decision}
	logs := &recordingLogProvider{}
	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{Decision: decisions, Log: logs},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	err := node.Execute(c)
	return c, decisions, logs.messages, err
}

func logNode(id string, message string) *CompiledFlowNode {
	return &CompiledFlowNode{ID: id, Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: message}}
}

func aiChoiceNode(minConfidence string, options ...string) *CompiledFlowNode {
	node := &CompiledFlowNode{
		ID:   "choice",
		Type: FlowNodeTypeControlConditionAIChoice,
		Data: FlowNodeData{AIDecisionData: &AIDecisionData{
			Question:      "What is the message about?",
			Input:         "How do I add a command?",
			MinConfidence: minConfidence,
		}},
	}
	branches := []*CompiledFlowNode{}
	for _, option := range options {
		branches = append(branches, &CompiledFlowNode{
			ID:       "option-" + option,
			Type:     FlowNodeTypeControlConditionItemAIChoice,
			Data:     FlowNodeData{AIChoiceName: option},
			Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{logNode("log-"+option, option)}},
		})
	}
	branches = append(branches, &CompiledFlowNode{
		ID:       "else",
		Type:     FlowNodeTypeControlConditionItemElse,
		Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{logNode("log-else", "else")}},
	})
	for _, branch := range branches {
		branch.Parents.Default = []*CompiledFlowNode{node}
	}
	node.Children.Default = branches
	return node
}

func TestAIChoiceRunsPickedOption(t *testing.T) {
	node := aiChoiceNode("0.8", "spam", "question")
	c, decisions, logs, err := executeDecisionNode(t, node, provider.Decision{Choice: "question", Confidence: 0.9})
	require.NoError(t, err)
	assert.Equal(t, []string{"question"}, logs)
	assert.Equal(t, "question", c.GetNodeResult("choice").String())

	assert.Equal(t, aiDecisionModel, decisions.opts.Model)
	assert.Equal(t, provider.DecisionTypeChoice, decisions.opts.Type)
	assert.Equal(t, map[string]string{"spam": "", "question": ""}, decisions.opts.Options)
	assert.Equal(t, "How do I add a command?", decisions.opts.Input)
}

func TestAIChoiceRunsElseBelowMinConfidence(t *testing.T) {
	node := aiChoiceNode("0.95", "spam", "question")
	c, _, logs, err := executeDecisionNode(t, node, provider.Decision{Choice: "question", Confidence: 0.9})
	require.NoError(t, err)
	assert.Equal(t, []string{"else"}, logs)
	assert.Equal(t, "question", c.GetNodeResult("choice").String())
}

func TestAIChoiceIgnoresUnknownChoice(t *testing.T) {
	node := aiChoiceNode("", "spam", "question")
	c, _, logs, err := executeDecisionNode(t, node, provider.Decision{Choice: "other", Confidence: 1})
	require.NoError(t, err)
	assert.Equal(t, []string{"else"}, logs)
	assert.Equal(t, "", c.GetNodeResult("choice").String())
}

func TestAIChoiceRejectsDuplicateOptions(t *testing.T) {
	node := aiChoiceNode("", "spam", "spam")
	_, decisions, _, err := executeDecisionNode(t, node, provider.Decision{Choice: "spam"})
	assert.ErrorContains(t, err, `option "spam" exists twice`)
	assert.Nil(t, decisions.opts)
}

func aiCheckNode(yesThreshold string, noThreshold string) *CompiledFlowNode {
	return &CompiledFlowNode{
		ID:   "check",
		Type: FlowNodeTypeControlAICheck,
		Data: FlowNodeData{AIDecisionData: &AIDecisionData{
			Question:     "Is this message spam?",
			Input:        "Free nitro!",
			YesThreshold: yesThreshold,
			NoThreshold:  noThreshold,
		}},
		Children: ConnectedFlowNodes{Handles: map[string][]*CompiledFlowNode{
			"yes":    {logNode("yes", "yes")},
			"no":     {logNode("no", "no")},
			"unsure": {logNode("unsure", "unsure")},
		}},
	}
}

func TestAICheckBranches(t *testing.T) {
	cases := []struct {
		yes, no string
		noul    float64
		want    string
	}{
		{"0.9", "0.5", 0.95, "yes"},
		{"0.9", "0.5", 0.9, "yes"},
		{"0.9", "0.5", 0.6, "unsure"},
		{"0.9", "0.5", 0.5, "unsure"},
		{"0.9", "0.5", 0.2, "no"},
		// Without thresholds there is no unsure.
		{"", "", 0.5, "yes"},
		{"", "", 0.49, "no"},
		// One threshold is used for both.
		{"0.3", "", 0.4, "yes"},
		{"0.3", "", 0.2, "no"},
		{"", "0.7", 0.6, "no"},
	}

	for _, tc := range cases {
		c, decisions, logs, err := executeDecisionNode(t, aiCheckNode(tc.yes, tc.no), provider.Decision{Noul: tc.noul})
		require.NoError(t, err)
		assert.Equal(t, []string{tc.want}, logs, "%v", tc)
		assert.Equal(t, tc.noul, c.GetNodeResult("check").Float())
		assert.Equal(t, provider.DecisionTypeNoul, decisions.opts.Type)
	}
}

func TestAICheckRejectsInvalidThresholds(t *testing.T) {
	for _, tc := range []struct{ yes, no, err string }{
		{"0.5", "0.6", "no threshold can't be above yes threshold"},
		{"abc", "", `yes threshold must be a number from 0 to 1, got "abc"`},
		{"0.5", "1.5", `no threshold must be a number from 0 to 1, got "1.5"`},
		{"NaN", "", `yes threshold must be a number from 0 to 1, got "NaN"`},
	} {
		_, decisions, _, err := executeDecisionNode(t, aiCheckNode(tc.yes, tc.no), provider.Decision{})
		assert.ErrorContains(t, err, tc.err)
		assert.Nil(t, decisions.opts)
	}
}

func TestAIDecisionCapsInput(t *testing.T) {
	node := aiCheckNode("", "")
	// Half the limit in characters, but é is 2 bytes.
	node.Data.AIDecisionData.Input = strings.Repeat("é", aiDecisionMaxInputLength/2)
	_, decisions, _, err := executeDecisionNode(t, node, provider.Decision{})
	assert.ErrorContains(t, err, "more than the limit of 8000")
	assert.Nil(t, decisions.opts)
}

func TestAIDecisionRejectsEmptyInput(t *testing.T) {
	node := aiCheckNode("", "")
	node.Data.AIDecisionData.Input = " "
	_, decisions, _, err := executeDecisionNode(t, node, provider.Decision{})
	assert.ErrorContains(t, err, "input is empty")
	assert.Nil(t, decisions.opts)
}

func TestAIScoreCountsLevelsFromOne(t *testing.T) {
	node := &CompiledFlowNode{
		ID:   "score",
		Type: FlowNodeTypeActionAIScore,
		Data: FlowNodeData{AIDecisionData: &AIDecisionData{
			Question: "How urgent is the ticket?",
			Input:    "The bot is down",
			Levels:   []string{"Not urgent", "Urgent", "Very urgent"},
		}},
	}
	c, decisions, _, err := executeDecisionNode(t, node, provider.Decision{Score: 1.4})
	require.NoError(t, err)
	assert.InDelta(t, 2.4, c.GetNodeResult("score").Float(), 1e-9)
	assert.Equal(t, provider.DecisionTypeScore, decisions.opts.Type)
	assert.Equal(t, []string{"Not urgent", "Urgent", "Very urgent"}, decisions.opts.Levels)

	node.Data.AIDecisionData.Levels = []string{"Only"}
	_, _, _, err = executeDecisionNode(t, node, provider.Decision{})
	assert.ErrorContains(t, err, "needs 2 to 10 levels, has 1")
}
