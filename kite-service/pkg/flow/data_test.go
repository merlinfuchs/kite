package flow

import (
	"context"
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAIModel(t *testing.T) {
	allowed := []string{
		"", // unset
		AIModelSmall,
		AIModelMedium,
		AIModelLarge,
		// Stored before tiers existed.
		"gpt-4.1",
		"gpt-4.1-mini",
		"gpt-4.1-nano",
		"gpt-4o-mini",
		"gpt-5-nano",
		"gpt-5.4-nano",
	}
	for _, model := range allowed {
		if _, ok := resolveAIModel(model); !ok {
			t.Errorf("model %q should be allowed", model)
		}
	}

	denied := []string{"o3", "gpt-4.5-preview", "gpt-5", "gpt-6-luna", "../etc/passwd"}
	for _, model := range denied {
		if _, ok := resolveAIModel(model); ok {
			t.Errorf("model %q should not be allowed", model)
		}
	}
}

// Moving to tiers must not change what an existing block is billed.
func TestAICreditsCostKnownModels(t *testing.T) {
	cases := []struct {
		model     string
		webSearch bool
		want      int
	}{
		{AIModelLarge, false, 100},
		{AIModelLarge, true, 500},
		{AIModelMedium, false, 20},
		{AIModelMedium, true, 100},
		{AIModelSmall, false, 5},
		{AIModelSmall, true, 25},
		{"gpt-4.1", false, 100},
		{"gpt-4.1", true, 500},
		{"gpt-4.1-mini", false, 20},
		{"gpt-4.1-mini", true, 100},
		{"gpt-4.1-nano", false, 5},
		{"gpt-4.1-nano", true, 25},
		{"gpt-4o-mini", false, 5},
		{"gpt-4o-mini", true, 25},
		{"gpt-5-nano", false, 5},
		{"gpt-5-nano", true, 25},
		{"gpt-5.4-nano", false, 5},
		{"gpt-5.4-nano", true, 25},
		{"", false, 5},
		{"", true, 25},
	}

	for _, tc := range cases {
		if got := AICreditsCost(tc.model, tc.webSearch); got != tc.want {
			t.Errorf("AICreditsCost(%q, %v) = %d, want %d",
				tc.model, tc.webSearch, got, tc.want)
		}
	}
}

// The old pricing switch had a cheap default arm, so a model it did not name
// billed at the floor while the operator paid the real rate. Unknown models
// have to price at the ceiling instead.
func TestAICreditsCostUnknownModelPricesAtCeiling(t *testing.T) {
	const unknown = "some-future-expensive-model"

	if _, ok := resolveAIModel(unknown); ok {
		t.Fatal("setup: model is supposed to be unknown")
	}

	chat := AICreditsCost(unknown, false)
	if want := AICreditsCost(AIModelLarge, false); chat != want {
		t.Errorf("unknown chat cost = %d, want the most expensive tier's %d", chat, want)
	}

	search := AICreditsCost(unknown, true)
	if want := AICreditsCost(AIModelLarge, true); search != want {
		t.Errorf("unknown search cost = %d, want the most expensive tier's %d", search, want)
	}
}

func TestAIModelAliasesPointAtTiers(t *testing.T) {
	for alias, tier := range aiModelAliases {
		assert.Contains(t, aiModelTiers, tier, "alias %q", alias)
		assert.NotContains(t, aiModelTiers, alias, "alias %q shadows a tier", alias)
	}
}

// The editor only offers what the catalog lists, so it has to match the tiers.
func TestCatalogAIModelsMatchTiers(t *testing.T) {
	catalog := loadCatalog(t)

	tiers := make([]any, 0, len(aiModelTiers))
	for tier := range aiModelTiers {
		tiers = append(tiers, tier)
	}

	for _, nodeType := range []FlowNodeType{
		FlowNodeTypeActionAIChatCompletion,
		FlowNodeTypeActionAISearchWeb,
	} {
		model := catalog[string(nodeType)].Properties["ai_chat_completion_data"].Properties["model"]
		assert.ElementsMatch(t, tiers, model.Enum, nodeType)
	}
}

func TestAIChatCompletionDataRejectsUnknownModel(t *testing.T) {
	err := AIChatCompletionData{Model: "o3", Prompt: "hi"}.Validate()
	if err == nil {
		t.Fatal("expected validation error for unknown model, got nil")
	}

	for _, model := range []string{AIModelLarge, "gpt-4.1"} {
		if err := (AIChatCompletionData{Model: model, Prompt: "hi"}).Validate(); err != nil {
			t.Fatalf("expected %q to validate, got %v", model, err)
		}
	}
}

// Both AI node types read AIChatCompletionData, so both have to require it.
func TestFlowNodeDataRequiresAIDataForBothAINodes(t *testing.T) {
	for _, nodeType := range []FlowNodeType{
		FlowNodeTypeActionAIChatCompletion,
		FlowNodeTypeActionAISearchWeb,
	} {
		if err := (FlowNodeData{}).Validate(nodeType); err == nil {
			t.Errorf("%s: expected error for missing ai data, got nil", nodeType)
		}
	}
}

func TestPollDataToCreatePollData(t *testing.T) {
	data := PollData{
		Question: "Best number? {{ 1 + 1 }}",
		Answers: []PollAnswerData{
			{Text: "One", Emoji: &EmojiData{Name: "1️⃣"}},
			{Text: "", Emoji: &EmojiData{Name: "2️⃣"}},
			{Text: "Custom", Emoji: &EmojiData{ID: "123456789012345678", Name: "custom"}},
		},
		DurationHours:    "{{ 24 * 2 }}",
		AllowMultiselect: true,
	}

	res, err := data.ToCreatePollData(context.Background(), eval.NewContext(eval.Env{}))
	require.NoError(t, err)

	assert.Equal(t, "Best number? 2", res.Question.Text)
	assert.Equal(t, 48, res.Duration)
	assert.True(t, res.AllowMultiselect)
	assert.Equal(t, provider.PollLayoutTypeDefault, res.LayoutType)

	// The empty answer is skipped.
	require.Len(t, res.Answers, 2)
	assert.Equal(t, "One", res.Answers[0].PollMedia.Text)
	assert.Equal(t, &provider.PollEmoji{Name: "1️⃣"}, res.Answers[0].PollMedia.Emoji)
	assert.Equal(t, &provider.PollEmoji{ID: discord.EmojiID(123456789012345678)}, res.Answers[1].PollMedia.Emoji)
}

func TestPollDataToCreatePollDataDefaultDuration(t *testing.T) {
	data := PollData{
		Question: "Question",
		Answers:  []PollAnswerData{{Text: "Answer"}},
	}

	res, err := data.ToCreatePollData(context.Background(), eval.NewContext(eval.Env{}))
	require.NoError(t, err)
	assert.Equal(t, pollDefaultDurationHours, res.Duration)
}

func TestPollDataToCreatePollDataRejectsInvalid(t *testing.T) {
	answers := func(n int) []PollAnswerData {
		res := make([]PollAnswerData, n)
		for i := range res {
			res[i] = PollAnswerData{Text: "Answer"}
		}
		return res
	}

	cases := map[string]PollData{
		"empty question":           {Question: "", Answers: answers(2)},
		"question too long":        {Question: strings.Repeat("a", pollQuestionMaxLength+1), Answers: answers(2)},
		"no answers":               {Question: "Question"},
		"only empty answers":       {Question: "Question", Answers: []PollAnswerData{{Text: " "}}},
		"too many answers":         {Question: "Question", Answers: answers(pollMaxAnswers + 1)},
		"answer too long":          {Question: "Question", Answers: []PollAnswerData{{Text: strings.Repeat("a", pollAnswerMaxLength+1)}}},
		"invalid emoji id":         {Question: "Question", Answers: []PollAnswerData{{Text: "Answer", Emoji: &EmojiData{ID: "abc"}}}},
		"duration not a number":    {Question: "Question", Answers: answers(2), DurationHours: "soon"},
		"duration evaluates empty": {Question: "Question", Answers: answers(2), DurationHours: "{{ '' }}"},
		"duration zero":            {Question: "Question", Answers: answers(2), DurationHours: "0"},
		"duration too long":        {Question: "Question", Answers: answers(2), DurationHours: "769"},
		"duration fractional":      {Question: "Question", Answers: answers(2), DurationHours: "1.5"},
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := data.ToCreatePollData(context.Background(), eval.NewContext(eval.Env{}))
			assert.Error(t, err)
		})
	}
}
