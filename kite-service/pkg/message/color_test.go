package message

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColorUnmarshal(t *testing.T) {
	valid := map[string]int{
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
			assert.Equal(t, Color{Value: want}, c)
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

	assert.Equal(t, Color{Value: 0xFF0000}, data.Embeds[0].Color)
	require.NotNil(t, data.Components[0].AccentColor)
	assert.Equal(t, Color{Value: 0x00FF00}, *data.Components[0].AccentColor)

	// Saved back as numbers.
	b, err := json.Marshal(data.Embeds[0])
	require.NoError(t, err)
	assert.JSONEq(t, `{"description":"hi","color":16711680}`, string(b))
}

func TestColorTemplate(t *testing.T) {
	var data MessageData
	require.NoError(t, json.Unmarshal([]byte(`{
		"flags": 32768,
		"embeds": [{"description": "hi", "color": " {{arg('color')}} "}],
		"components": [
			{"type": 17, "accent_color": "{{arg('color')}}", "components": [{"type": 10, "content": "hi"}]},
			{"type": 17, "accent_color": "{{arg('none')}}", "components": [{"type": 10, "content": "hi"}]}
		]
	}`), &data))
	assert.Equal(t, Color{Template: "{{arg('color')}}"}, data.Embeds[0].Color)

	// Saved as written.
	b, err := json.Marshal(data.Embeds[0])
	require.NoError(t, err)
	assert.JSONEq(t, `{"description":"hi","color":"{{arg('color')}}"}`, string(b))

	// Sent without a flow to fill them in, they're left out.
	embed := data.Embeds[0].ToEmbed(time.Now())
	assert.Equal(t, discord.Color(0), embed.Color)
	container := data.Components[0].ToComponent(ConvertOptions{}).(*discord.ContainerComponent)
	assert.Nil(t, container.AccentColor)

	filled := data.Copy()
	require.NoError(t, filled.FillColors(func(template string) (string, error) {
		if template == "{{arg('color')}}" {
			return "#00ff00", nil
		}
		return "", nil
	}))
	assert.Equal(t, Color{Value: 0x00FF00}, filled.Embeds[0].Color)
	require.NotNil(t, filled.Components[0].AccentColor)
	assert.Equal(t, Color{Value: 0x00FF00}, *filled.Components[0].AccentColor)
	assert.Nil(t, filled.Components[1].AccentColor)

	invalid := data.Copy()
	assert.Error(t, invalid.FillColors(func(string) (string, error) { return "red", nil }))
}
