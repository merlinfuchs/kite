package flowai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeOpenAI answers every request with a response of the given status and
// output text, and records the request body.
func fakeOpenAI(t *testing.T, status string, text string) (*Assistant, *map[string]any) {
	t.Helper()

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/responses", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &body))

		output, _ := json.Marshal(text)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"id": "resp_1", "object": "response", "created_at": 0, "status": %q, "model": "gpt-5-mini",
			"output": [{"type": "message", "id": "msg_1", "status": "completed", "role": "assistant",
				"content": [{"type": "output_text", "text": %s, "annotations": []}]}],
			"usage": {"input_tokens": 1000, "input_tokens_details": {"cached_tokens": 800},
				"output_tokens": 200, "output_tokens_details": {"reasoning_tokens": 100}, "total_tokens": 1200}
		}`, status, output)
	}))
	t.Cleanup(server.Close)

	client := openai.NewClient(option.WithAPIKey("test"), option.WithBaseURL(server.URL))
	assistant := NewAssistant(&client, Config{Model: "gpt-5-mini", ReasoningEffort: "low", MaxOutputTokens: 1000})
	return assistant, &body
}

func TestRespond(t *testing.T) {
	assistant, body := fakeOpenAI(t, "completed", `{"message": "Added a log.", "edits": [
		{"op": "add_node", "ref": "$log", "type": "action_log", "id": null, "after": "entry", "before": null,
		 "handle": null, "source": null, "target": null, "data_json": "{\"log_level\": \"info\", \"log_message\": \"hi\"}",
		 "items_json": null, "reconnect": null}
	]}`)

	messages := []Message{}
	for i := range 12 {
		messages = append(messages, Message{Role: "user", Content: fmt.Sprintf("old %d", i)})
	}
	messages = append(messages, Message{Role: "user", Content: "Add a log"})

	res, err := assistant.Respond(context.Background(), Request{
		FlowType: "command",
		Flow:     "Blocks:\n- entry entry_command",
		Messages: messages,
		UserID:   "user",
	})
	require.NoError(t, err)

	assert.Equal(t, "Added a log.", res.Message)
	assert.Equal(t, []map[string]any{{
		"op":    "add_node",
		"ref":   "$log",
		"type":  "action_log",
		"after": "entry",
		"data":  map[string]any{"log_level": "info", "log_message": "hi"},
	}}, res.Edits)
	assert.Empty(t, res.Issues)
	assert.Equal(t, model.FlowAIUsage{InputTokens: 1000, CachedInputTokens: 800, OutputTokens: 200}, res.Usage)

	req := *body
	assert.Equal(t, "gpt-5-mini", req["model"])
	assert.Equal(t, "kite-flow-ai", req["prompt_cache_key"])
	assert.Contains(t, req["instructions"], "Block catalog:")
	format := req["text"].(map[string]any)["format"].(map[string]any)
	assert.Equal(t, "json_schema", format["type"])
	assert.Equal(t, true, format["strict"])

	// Of the 12 earlier messages, the oldest 5 are dropped, then the current
	// one is sent with the flow.
	input := req["input"].([]any)
	require.Len(t, input, 8)
	assert.Equal(t, "old 5", input[0].(map[string]any)["content"])
	last := input[7].(map[string]any)["content"].(string)
	assert.Contains(t, last, "Flow type: command")
	assert.Contains(t, last, "- entry entry_command")
	assert.True(t, strings.HasSuffix(last, "Add a log"))
}

func TestRespondRepair(t *testing.T) {
	assistant, body := fakeOpenAI(t, "completed", `{"message": "Fixed.", "edits": []}`)

	_, err := assistant.Respond(context.Background(), Request{
		FlowType: "command",
		Flow:     "Blocks:",
		Messages: []Message{
			{Role: "user", Content: "Add a log"},
			{Role: "assistant", Content: "Added a log."},
		},
		Issues: []string{"'Log Message' setting 'log_level': Required"},
	})
	require.NoError(t, err)

	input := (*body)["input"].([]any)
	require.Len(t, input, 3)
	assert.Equal(t, "assistant", input[1].(map[string]any)["role"])
	assert.Contains(t, input[2].(map[string]any)["content"], "- 'Log Message' setting 'log_level': Required")
}

func TestRespondCutOff(t *testing.T) {
	assistant, _ := fakeOpenAI(t, "incomplete", `{"message": "Added`)

	_, err := assistant.Respond(context.Background(), Request{
		FlowType: "command",
		Flow:     "Blocks:",
		Messages: []Message{{Role: "user", Content: "Add a log"}},
	})

	var resErr *ErrResponse
	assert.True(t, errors.As(err, &resErr))
}

func TestParseOutputSkipsEditsWithInvalidSettings(t *testing.T) {
	res, err := parseOutput(`{"message": "", "edits": [
		{"op": "update_node", "ref": null, "type": null, "id": "a", "after": null, "before": null, "handle": null,
		 "source": null, "target": null, "data_json": "not json", "items_json": null, "reconnect": null},
		{"op": "add_node", "ref": "$c", "type": "control_condition_compare", "id": null, "after": "entry",
		 "before": null, "handle": null, "source": null, "target": null, "data_json": "{}",
		 "items_json": "[{\"condition_item_mode\": \"equal\"}]", "reconnect": null},
		{"op": "remove_node", "ref": null, "type": null, "id": "b", "after": null, "before": null, "handle": null,
		 "source": null, "target": null, "data_json": null, "items_json": null, "reconnect": false}
	]}`)
	require.NoError(t, err)

	assert.Equal(t, []string{"Edit 1 (update_node) was skipped: data_json isn't a JSON object"}, res.Issues)
	require.Len(t, res.Edits, 2)
	assert.Equal(t, []map[string]any{{"condition_item_mode": "equal"}}, res.Edits[0]["items"])
	assert.Equal(t, map[string]any{"op": "remove_node", "id": "b", "reconnect": false}, res.Edits[1])
}
