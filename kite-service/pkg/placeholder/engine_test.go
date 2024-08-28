package placeholder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testCases = []struct {
	input    string
	expected string
}{
	{
		input:    "Hello, {{string}}!",
		expected: "Hello, World!",
	},
	{
		input:    "Hello, {{map.string}}!",
		expected: "Hello, Worlds!",
	},
	{
		input:    "Hello, {{ map.string }}!",
		expected: "Hello, Worlds!",
	},
	{
		input:    "Hello, {{invalid}}!",
		expected: "Hello, !",
	},
}

func TestFlowExecuteCommand(t *testing.T) {
	engine := NewEngine()
	engine.AddProvider("string", NewStringProvider("World"))
	engine.AddProvider("map", NewMapProvider(map[string]Provider{
		"string": NewStringProvider("Worlds"),
	}))

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual, err := engine.Fill(context.TODO(), tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
