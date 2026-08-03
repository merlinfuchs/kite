package flow

import (
	"testing"

	"github.com/openai/openai-go"
)

func TestAIModelAllowed(t *testing.T) {
	allowed := []string{
		"", // provider default
		openai.ChatModelGPT4_1,
		openai.ChatModelGPT4_1Mini,
		openai.ChatModelGPT4_1Nano,
		openai.ChatModelGPT4oMini,
	}
	for _, model := range allowed {
		if !AIModelAllowed(model) {
			t.Errorf("model %q should be allowed", model)
		}
	}

	denied := []string{"o3", "gpt-4.5-preview", "gpt-5", "../etc/passwd"}
	for _, model := range denied {
		if AIModelAllowed(model) {
			t.Errorf("model %q should not be allowed", model)
		}
	}
}

func TestAICreditsCostKnownModels(t *testing.T) {
	cases := []struct {
		model     string
		webSearch bool
		want      int
	}{
		{openai.ChatModelGPT4_1, false, 100},
		{openai.ChatModelGPT4_1, true, 500},
		{openai.ChatModelGPT4_1Mini, false, 20},
		{openai.ChatModelGPT4_1Mini, true, 100},
		{openai.ChatModelGPT4oMini, false, 5},
		{openai.ChatModelGPT4oMini, true, 25},
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

	if AIModelAllowed(unknown) {
		t.Fatal("setup: model is supposed to be unknown")
	}

	chat := AICreditsCost(unknown, false)
	if want := AICreditsCost(openai.ChatModelGPT4_1, false); chat != want {
		t.Errorf("unknown chat cost = %d, want the most expensive model's %d", chat, want)
	}

	search := AICreditsCost(unknown, true)
	if want := AICreditsCost(openai.ChatModelGPT4_1, true); search != want {
		t.Errorf("unknown search cost = %d, want the most expensive model's %d", search, want)
	}
}

func TestAIChatCompletionDataRejectsUnknownModel(t *testing.T) {
	err := AIChatCompletionData{Model: "o3", Prompt: "hi"}.Validate()
	if err == nil {
		t.Fatal("expected validation error for unknown model, got nil")
	}

	if err := (AIChatCompletionData{Model: openai.ChatModelGPT4_1, Prompt: "hi"}).Validate(); err != nil {
		t.Fatalf("expected allowed model to validate, got %v", err)
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
