package message

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColorUnmarshal(t *testing.T) {
	valid := map[string]Color{
		`5793266`:     0x5865F2,
		`5793266.0`:   0x5865F2,
		`null`:        0,
		`"#5865F2"`:   0x5865F2,
		`"#5865f2"`:   0x5865F2,
		`"0x5865F2"`:  0x5865F2,
		`"5793266"`:   0x5865F2,
		`" #ff0000 "`: 0xFF0000,
		`"#fff"`:      0xFFFFFF,
		`"#f00"`:      0xFF0000,
		// Stored numbers load even when Discord would reject them.
		`16777216`:   0x1000000,
		`4294967295`: 0xFFFFFFFF,
	}
	for raw, want := range valid {
		t.Run(raw, func(t *testing.T) {
			var c Color
			require.NoError(t, json.Unmarshal([]byte(raw), &c))
			assert.Equal(t, want, c)
		})
	}

	invalid := []string{
		`""`,
		`"red"`,
		`"ff0000"`,
		`"#zzzzzz"`,
		`"#5865F2FF"`,
		`"#-1"`,
		`"-1"`,
		`"16777216"`,
		`1.5`,
		`1e19`,
		`true`,
		`{}`,
	}
	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			var c Color
			assert.Error(t, json.Unmarshal([]byte(raw), &c))
		})
	}
}

func TestEmbedColorFromText(t *testing.T) {
	var data MessageData
	require.NoError(t, json.Unmarshal([]byte(`{
		"embeds": [{"description": "hi", "color": "#ff0000"}],
		"components": [{"type": 17, "accent_color": "#00ff00"}]
	}`), &data))

	assert.Equal(t, Color(0xFF0000), data.Embeds[0].Color)
	require.NotNil(t, data.Components[0].AccentColor)
	assert.Equal(t, Color(0x00FF00), *data.Components[0].AccentColor)

	// Saved back as numbers.
	b, err := json.Marshal(data.Embeds[0])
	require.NoError(t, err)
	assert.JSONEq(t, `{"description":"hi","color":16711680}`, string(b))
}
