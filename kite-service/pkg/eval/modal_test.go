package eval

import (
	"encoding/json"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
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
	assert.Equal(t, "red", components["colors"].Value)
	assert.Equal(t, []string{"red", "blue"}, components["colors"].Values)
	assert.Equal(t, "123", components["users"].Value)
	assert.Equal(t, "large", components["size"].Value)
	assert.Equal(t, "", components["extras"].Value)
	assert.Empty(t, components["extras"].Values)
	assert.Equal(t, "true", components["agree"].Value)
}
