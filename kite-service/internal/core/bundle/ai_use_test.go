package bundle

import (
	"encoding/json"
	"testing"
)

func TestAIUse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "generated at the 80 percent threshold",
			content: `{"commands":[{"flow_source":{"nodes":[
				{"type":"entry_command","data":{}},
				{"type":"action_a","data":{"ai_contribution":"generated"}},
				{"type":"action_b","data":{"ai_contribution":"generated"}},
				{"type":"action_c","data":{"ai_contribution":"generated"}},
				{"type":"action_d","data":{"ai_contribution":"generated"}}
			]}}]}`,
			want: "generated",
		},
		{
			name: "assisted below the 80 percent threshold",
			content: `{"commands":[{"flow_source":{"nodes":[
				{"type":"entry_command","data":{}},
				{"type":"action_a","data":{"ai_contribution":"generated"}},
				{"type":"action_b","data":{}}
			]}}]}`,
			want: "assisted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AIUse(json.RawMessage(tt.content)); got != tt.want {
				t.Fatalf("AIUse() = %q, want %q", got, tt.want)
			}
		})
	}
}
