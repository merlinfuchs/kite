package eval

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComponentsEnvModal(t *testing.T) {
	// A modal submission as Discord sends it, with a legacy action row next
	// to labels holding the newer inputs.
	data := `{
		"custom_id": "modal",
		"components": [
			{"type": 1, "components": [{"type": 4, "custom_id": "old", "value": "legacy"}]},
			{"type": 10, "id": 2, "content": "Hello"},
			{"type": 18, "id": 3, "component": {"type": 4, "custom_id": "name", "value": "Jack"}},
			{"type": 18, "id": 4, "component": {"type": 3, "custom_id": "colors", "values": ["red", "blue"]}},
			{"type": 18, "id": 5, "component": {"type": 5, "custom_id": "users", "values": ["123"]}},
			{"type": 18, "id": 6, "component": {"type": 21, "custom_id": "size", "value": "large"}},
			{"type": 18, "id": 7, "component": {"type": 22, "custom_id": "extras", "values": []}},
			{"type": 18, "id": 8, "component": {"type": 23, "custom_id": "agree", "value": true}}
		]
	}`

	var modal discord.ModalInteraction
	require.NoError(t, json.Unmarshal([]byte(data), &modal))

	components := NewComponentsEnv(&discord.InteractionEvent{Data: &modal})

	assert.Equal(t, "legacy", components["old"].Value)
	assert.Equal(t, "Jack", components["name"].Value)
	assert.Equal(t, "red, blue", components["colors"].Value)
	assert.Equal(t, []string{"red", "blue"}, components["colors"].Values)
	assert.Equal(t, "123", components["users"].Value)
	assert.Equal(t, "large", components["size"].Value)
	assert.Equal(t, "", components["extras"].Value)
	assert.Empty(t, components["extras"].Values)
	assert.Equal(t, "true", components["agree"].Value)
}

func TestModalInputValues(t *testing.T) {
	data := `{
		"custom_id": "modal",
		"components": [
			{"type": 18, "id": 1, "component": {"type": 3, "custom_id": "color", "values": ["red"]}},
			{"type": 18, "id": 2, "component": {"type": 22, "custom_id": "extras", "values": ["cheese", "bacon"]}},
			{"type": 18, "id": 3, "component": {"type": 22, "custom_id": "none", "values": []}}
		]
	}`

	var modal discord.ModalInteraction
	require.NoError(t, json.Unmarshal([]byte(data), &modal))

	c := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &modal,
	}, state.New("Bot test"))

	tests := map[string]string{
		// A single pick is returned as is.
		"{{input('color')}}": "red",
		// Several picks are joined, so none is dropped.
		"{{input('extras')}}":                                    "cheese, bacon",
		"Extras: {{input('extras')}}":                            "Extras: cheese, bacon",
		"{{input('none')}}":                                      "",
		"{{interaction.components['extras'].values[1]}}":         "bacon",
		"{{len(interaction.components['extras'].values)}}":       "2",
		"{{'bacon' in interaction.components['extras'].values}}": "true",
		"{{'bac' in interaction.components['extras'].values}}":   "false",
	}
	for template, want := range tests {
		got, err := EvalTemplateToString(context.Background(), template, c)
		require.NoError(t, err, template)
		assert.Equal(t, want, got, template)
	}
}
