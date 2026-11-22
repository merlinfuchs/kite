package flowtest

import (
	"encoding/json"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func TestVariableValue(t *testing.T) {
	tests := []struct {
		name    string
		v       wire.FlowTestVariable
		want    string
		wantErr bool
	}{
		{"json", wire.FlowTestVariable{Value: json.RawMessage(`{"xp": 3}`)}, `{"xp":3}`, false},
		{"no value", wire.FlowTestVariable{}, `null`, false},
		{"number", wire.FlowTestVariable{Type: wire.VariableValueTypeNumber, Value: json.RawMessage(`"5"`)}, `5`, false},
		{"text keeps zeros", wire.FlowTestVariable{Type: wire.VariableValueTypeString, Value: json.RawMessage(`"007"`)}, `"007"`, false},
		{"invalid number", wire.FlowTestVariable{Type: wire.VariableValueTypeNumber, Value: json.RawMessage(`"10k"`)}, "", true},
		{"invalid bool", wire.FlowTestVariable{Type: wire.VariableValueTypeBool, Value: json.RawMessage(`"yes"`)}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := testVariableValue(tt.v)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("got %v, want an error", v)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got, _ := json.Marshal(v.JSONValue())
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
