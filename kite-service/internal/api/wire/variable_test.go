package wire

import (
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariableValueSetRequestThing(t *testing.T) {
	tests := []struct {
		name    string
		req     VariableValueSetRequest
		want    thing.Thing
		wantErr bool
	}{
		{"text", VariableValueSetRequest{Type: VariableValueTypeString, Value: " 12 "}, thing.NewString(" 12 "), false},
		{"empty text", VariableValueSetRequest{Type: VariableValueTypeString}, thing.NewString(""), false},
		{"integer", VariableValueSetRequest{Type: VariableValueTypeNumber, Value: " -42 "}, thing.NewInt(-42), false},
		{"large integer", VariableValueSetRequest{Type: VariableValueTypeNumber, Value: "1497746534387941386"}, thing.NewInt(1497746534387941386), false},
		{"decimal", VariableValueSetRequest{Type: VariableValueTypeNumber, Value: "1.5"}, thing.NewFloat(1.5), false},
		{"empty number", VariableValueSetRequest{Type: VariableValueTypeNumber}, thing.Null, true},
		{"invalid number", VariableValueSetRequest{Type: VariableValueTypeNumber, Value: "12abc"}, thing.Null, true},
		{"infinite number", VariableValueSetRequest{Type: VariableValueTypeNumber, Value: "Inf"}, thing.Null, true},
		{"true", VariableValueSetRequest{Type: VariableValueTypeBool, Value: "true"}, thing.NewBool(true), false},
		{"invalid bool", VariableValueSetRequest{Type: VariableValueTypeBool, Value: "yes"}, thing.Null, true},
		{"json null", VariableValueSetRequest{Type: VariableValueTypeJSON, Value: "null"}, thing.Null, false},
		{
			"json object",
			VariableValueSetRequest{Type: VariableValueTypeJSON, Value: `{"id": 1497746534387941386, "tags": ["a", 1.5, true]}`},
			thing.NewObject(map[string]thing.Thing{
				"id": thing.NewInt(1497746534387941386),
				"tags": thing.NewArray([]thing.Thing{
					thing.NewString("a"),
					thing.NewFloat(1.5),
					thing.NewBool(true),
				}),
			}),
			false,
		},
		{"invalid json", VariableValueSetRequest{Type: VariableValueTypeJSON, Value: `{"a":`}, thing.Null, true},
		{"trailing json", VariableValueSetRequest{Type: VariableValueTypeJSON, Value: `1 2`}, thing.Null, true},
		{"unknown type", VariableValueSetRequest{Type: "other", Value: "1"}, thing.Null, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.req.Thing()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVariableValueSetRequestValidate(t *testing.T) {
	assert.NoError(t, VariableValueSetRequest{Type: VariableValueTypeString}.Validate())
	assert.Error(t, VariableValueSetRequest{Value: "a"}.Validate())
	assert.Error(t, VariableValueSetRequest{Type: "other", Value: "a"}.Validate())
	assert.Error(t, VariableValueSetRequest{
		Type:  VariableValueTypeString,
		Value: strings.Repeat("a", MaxVariableValueLength+1),
	}.Validate())
	assert.Error(t, VariableValueSetRequest{
		Scope: strings.Repeat("a", MaxVariableScopeLength+1),
		Type:  VariableValueTypeString,
	}.Validate())
}

func TestVariableValueToWire(t *testing.T) {
	toWire := func(data thing.Thing) *VariableValue {
		return VariableValueToWire(&model.VariableValue{Data: data})
	}

	v := toWire(thing.NewString("<b>hi</b>"))
	assert.Equal(t, VariableValueTypeString, v.Type)
	assert.Equal(t, "<b>hi</b>", v.Value)
	assert.False(t, v.ReadOnly)

	v = toWire(thing.NewInt(1497746534387941386))
	assert.Equal(t, VariableValueTypeNumber, v.Type)
	assert.Equal(t, "1497746534387941386", v.Value)

	v = toWire(thing.NewFloat(1.5))
	assert.Equal(t, VariableValueTypeNumber, v.Type)
	assert.Equal(t, "1.5", v.Value)

	v = toWire(thing.NewBool(false))
	assert.Equal(t, VariableValueTypeBool, v.Type)
	assert.Equal(t, "false", v.Value)

	v = toWire(thing.Null)
	assert.Equal(t, VariableValueTypeJSON, v.Type)
	assert.Equal(t, "null", v.Value)
	assert.False(t, v.ReadOnly)

	v = toWire(thing.NewArray([]thing.Thing{thing.NewInt(1), thing.NewString("<a>")}))
	assert.Equal(t, VariableValueTypeJSON, v.Type)
	assert.Equal(t, "[\n  1,\n  \"<a>\"\n]", v.Value)
	assert.False(t, v.ReadOnly)

	// Discord objects can't be entered as text, so they are read-only, also
	// when nested.
	v = toWire(thing.NewDiscordUser(discord.User{ID: 123, Username: "kite"}))
	assert.Equal(t, VariableValueTypeJSON, v.Type)
	assert.Contains(t, v.Value, `"username": "kite"`)
	assert.True(t, v.ReadOnly)

	v = toWire(thing.NewArray([]thing.Thing{thing.NewDiscordUser(discord.User{ID: 123})}))
	assert.True(t, v.ReadOnly)

	v = toWire(thing.NewString(strings.Repeat("ä", MaxVariableValuePreviewLength)))
	assert.True(t, v.Truncated)
	assert.LessOrEqual(t, len(v.Value), MaxVariableValuePreviewLength)
	assert.Equal(t, strings.Repeat("ä", MaxVariableValuePreviewLength/2), v.Value)
}

// What the dashboard shows for a value must parse back to the same value.
func TestVariableValueRoundTrip(t *testing.T) {
	values := []thing.Thing{
		thing.NewString("hello\nworld"),
		thing.NewInt(-5),
		thing.NewFloat(0.25),
		thing.NewBool(true),
		thing.Null,
		thing.NewObject(map[string]thing.Thing{
			"list": thing.NewArray([]thing.Thing{thing.NewInt(1), thing.Null}),
		}),
	}

	for _, value := range values {
		shown := VariableValueToWire(&model.VariableValue{Data: value})
		got, err := VariableValueSetRequest{Type: shown.Type, Value: shown.Value}.Thing()
		require.NoError(t, err)
		assert.Equal(t, value, got)
	}
}
