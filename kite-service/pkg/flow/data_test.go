package flow

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIModelAllowed(t *testing.T) {
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
	}
	for _, model := range allowed {
		if !AIModelAllowed(model) {
			t.Errorf("model %q should be allowed", model)
		}
	}

	denied := []string{"o3", "gpt-4.5-preview", "gpt-5", "gpt-6-luna", "../etc/passwd"}
	for _, model := range denied {
		if AIModelAllowed(model) {
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
	var catalog struct {
		Nodes map[string]struct {
			DataSchema struct {
				Properties struct {
					AIChatCompletionData struct {
						Properties struct {
							Model struct {
								Enum []string `json:"enum"`
							} `json:"model"`
						} `json:"properties"`
					} `json:"ai_chat_completion_data"`
				} `json:"properties"`
			} `json:"data_schema"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal(CatalogJSON, &catalog))

	tiers := make([]string, 0, len(aiModelTiers))
	for tier := range aiModelTiers {
		tiers = append(tiers, tier)
	}

	for _, nodeType := range []FlowNodeType{
		FlowNodeTypeActionAIChatCompletion,
		FlowNodeTypeActionAISearchWeb,
	} {
		node, ok := catalog.Nodes[string(nodeType)]
		require.True(t, ok, nodeType)
		assert.ElementsMatch(t, tiers, node.DataSchema.Properties.AIChatCompletionData.Properties.Model.Enum, nodeType)
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
