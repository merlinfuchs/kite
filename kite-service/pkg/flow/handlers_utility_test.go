package flow

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeRandomID(t *testing.T, data FlowNodeData) (string, error) {
	c := NewContext(
		context.Background(),
		5*time.Second,
		&TestContextData{},
		FlowProviders{},
		FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 100,
			MaxCredits:    100,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	node := &CompiledFlowNode{ID: "0", Type: FlowNodeTypeActionRandomID, Data: data}
	if err := node.Execute(c); err != nil {
		return "", err
	}

	res := c.GetNodeResult("0")
	require.NotNil(t, res)
	return res.String(), nil
}

func TestFlowExecuteActionRandomID(t *testing.T) {
	tests := []struct {
		name    string
		idType  RandomIDType
		length  string
		pattern string
	}{
		{"default is uuid_v4", "", "", `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`},
		{"uuid_v4 ignores length", RandomIDTypeUUIDv4, "5", `^[0-9a-f-]{36}$`},
		{"nanoid default length", RandomIDTypeNanoID, "", `^[A-Za-z0-9_-]{21}$`},
		{"nanoid custom length", RandomIDTypeNanoID, "10", `^[A-Za-z0-9_-]{10}$`},
		{"alphanumeric default length", RandomIDTypeAlphanumeric, "", `^[A-Za-z0-9]{16}$`},
		{"alphanumeric custom length", RandomIDTypeAlphanumeric, "12", `^[A-Za-z0-9]{12}$`},
		{"numeric default length", RandomIDTypeNumeric, "", `^[0-9]{6}$`},
		{"numeric max length", RandomIDTypeNumeric, " 256 ", `^[0-9]{256}$`},
		{"hex default length", RandomIDTypeHex, "", `^[0-9a-f]{32}$`},
		{"hex templated length", RandomIDTypeHex, "{{ 8 * 2 }}", `^[0-9a-f]{16}$`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := executeRandomID(t, FlowNodeData{RandomIDType: tt.idType, RandomIDLength: tt.length})
			require.NoError(t, err)
			assert.Regexp(t, regexp.MustCompile(tt.pattern), id)
		})
	}
}

func TestFlowExecuteActionRandomIDInvalid(t *testing.T) {
	tests := []struct {
		name   string
		idType RandomIDType
		length string
	}{
		{"unknown type", "base64", ""},
		{"zero length", RandomIDTypeNumeric, "0"},
		{"negative length", RandomIDTypeNumeric, "-5"},
		{"too long", RandomIDTypeHex, "257"},
		{"not a number", RandomIDTypeNanoID, "ten"},
		{"fraction", RandomIDTypeNanoID, "1.5"},
		{"templated to empty", RandomIDTypeAlphanumeric, "{{ '' }}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeRandomID(t, FlowNodeData{RandomIDType: tt.idType, RandomIDLength: tt.length})
			assert.Error(t, err)
		})
	}
}
