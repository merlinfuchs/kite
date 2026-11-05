package flowai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalModelJSONRepairs(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  map[string]any
	}{
		{
			"missing closing bracket",
			`{"user_target":"1","message_data":{"content":"Hello!"}`,
			map[string]any{"user_target": "1", "message_data": map[string]any{"content": "Hello!"}},
		},
		{
			"missing ] before }",
			`{"message_data":{"embeds":[{"title":"a"}}}`,
			map[string]any{"message_data": map[string]any{"embeds": []any{map[string]any{"title": "a"}}}},
		},
		{
			"missing } before ]",
			`{"fields":[{"name":"a","value":"b"]}`,
			map[string]any{"fields": []any{map[string]any{"name": "a", "value": "b"}}},
		},
		{
			"extra closing bracket",
			`{"content":"a"}}`,
			map[string]any{"content": "a"},
		},
		{
			"raw line breaks and tabs",
			"{\"expression\":\"let a = 1;\n\tlet b = 2;\r\na + b\"}",
			map[string]any{"expression": "let a = 1;\n\tlet b = 2;\r\na + b"},
		},
		{
			"escaped single quote",
			`{"content":"by \'{{message.content}}\'"}`,
			map[string]any{"content": "by '{{message.content}}'"},
		},
		{
			"invalid escape",
			`{"content":"costs \$5"}`,
			map[string]any{"content": `costs \$5`},
		},
		{
			"brackets in strings",
			`{"content":"[{ok}"`,
			map[string]any{"content": "[{ok}"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]any
			require.NoError(t, unmarshalModelJSON(tt.value, &got))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnmarshalModelJSONKeepsOriginalError(t *testing.T) {
	var got map[string]any
	err := unmarshalModelJSON(`{"content":"say "hi""}`, &got)
	assert.EqualError(t, err, "invalid character 'h' after object key:value pair")
	assert.Nil(t, got)
}

func TestRepairJSONLeavesValidJSON(t *testing.T) {
	_, changed := repairJSON(`{"a":["b\n\"c\\"],"d":{}}`)
	assert.False(t, changed)
}
