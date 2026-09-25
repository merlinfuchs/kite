package flow

import (
	"context"
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

func executeAINode(t *testing.T, nodeType FlowNodeType, data AIChatCompletionData) provider.CreateResponseOpts {
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
	require.NoError(t, node.Execute(c))
	return ai.opts
}

func TestAINodeRunsResolvedTier(t *testing.T) {
	cases := []struct {
		model     string
		maxTokens string
		want      aiModelTier
		wantMax   int
	}{
		{"", "", aiModelTiers[AIModelSmall], aiMaxAnswerTokens},
		{"gpt-4.1", "100", aiModelTiers[AIModelLarge], 100},
		{AIModelMedium, "5000", aiModelTiers[AIModelMedium], aiMaxAnswerTokens + aiModelTiers[AIModelMedium].ReasoningTokens},
	}

	for _, tc := range cases {
		opts := executeAINode(t, FlowNodeTypeActionAIChatCompletion, AIChatCompletionData{
			Model:               tc.model,
			Prompt:              "hi",
			MaxCompletionTokens: tc.maxTokens,
		})
		assert.Equal(t, tc.want.Model, opts.Model, tc.model)
		assert.Equal(t, tc.want.ReasoningEffort, opts.ReasoningEffort, tc.model)
		assert.Equal(t, tc.wantMax, opts.MaxOutputTokens, tc.model)
		assert.Empty(t, opts.Tools, tc.model)
	}
}

func TestAIWebSearchNodeCapsSearches(t *testing.T) {
	opts := executeAINode(t, FlowNodeTypeActionAISearchWeb, AIChatCompletionData{Prompt: "hi"})
	assert.Equal(t, []provider.AIToolType{provider.AIToolTypeWebSearch}, opts.Tools)
	assert.Equal(t, aiMaxWebSearches, opts.MaxToolCalls)
}
