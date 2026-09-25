package flowai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	assistant, body := fakeOpenAI(t, "completed", `{"verdict": "clarify", "message": "Which channel?",
		"suggested_prompt": "Log bans in the channel below", "fields": [
		{"label": "Log channel", "description": "", "type": "channel", "options": [], "default": ""}]}`)

	res, err := assistant.Check(context.Background(), CheckRequest{
		Flow:   "Flow type: command\n\nBlocks:",
		Prompt: "Log bans",
		UserID: "user",
	})
	require.NoError(t, err)

	assert.Equal(t, "clarify", res.Verdict)
	assert.Equal(t, "Log bans in the channel below", res.SuggestedPrompt)
	assert.Equal(t, []CheckField{{Label: "Log channel", Type: "channel", Options: []string{}}}, res.Fields)

	req := *body
	assert.Equal(t, "gpt-5-nano", req["model"])
	assert.Equal(t, "minimal", req["reasoning"].(map[string]any)["effort"])
	assert.Contains(t, req["instructions"], "- Ban member: Ban a member from the server. Needs user_target (ID of the user)")
	assert.Contains(t, req["input"], "Log bans")
}

func TestCheckFailsIfCutOff(t *testing.T) {
	assistant, _ := fakeOpenAI(t, "incomplete", `{"verdict": "cl`)

	_, err := assistant.Check(context.Background(), CheckRequest{Flow: "Blocks:", Prompt: "Hi"})
	assert.Error(t, err)
}

func TestCheckLimitsFields(t *testing.T) {
	field := `{"label": "A", "description": "", "type": "choice", "options": [], "default": ""}`
	assistant, _ := fakeOpenAI(t, "completed", `{"verdict": "clarify", "message": "", "suggested_prompt": "",
		"fields": [`+field+`, `+field+`, `+field+`, `+field+`, `+field+`]}`)

	res, err := assistant.Check(context.Background(), CheckRequest{Flow: "Blocks:", Prompt: "Hi"})
	require.NoError(t, err)

	assert.Len(t, res.Fields, 4)
	assert.Equal(t, "text", res.Fields[0].Type)
	assert.Equal(t, 1000, res.Usage.InputTokens)
}
