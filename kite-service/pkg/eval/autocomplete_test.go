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

func TestAutocompleteContext(t *testing.T) {
	// The user types "app" in the fruit argument of /shop buy, after picking
	// an amount and a user.
	data := `{
		"id": "10",
		"name": "shop",
		"type": 1,
		"options": [{
			"type": 1,
			"name": "buy",
			"options": [
				{"type": 4, "name": "amount", "value": 3},
				{"type": 6, "name": "for", "value": "42"},
				{"type": 3, "name": "fruit", "value": "app", "focused": true}
			]
		}]
	}`

	var autocomplete discord.AutocompleteInteraction
	require.NoError(t, json.Unmarshal([]byte(data), &autocomplete))

	focused := FocusedOption(autocomplete.Options)
	assert.Equal(t, "fruit", focused.Name)
	assert.Equal(t, "app", focused.String())

	c := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &autocomplete,
	}, state.New("Bot test"))

	tests := map[string]string{
		"{{autocomplete.value}}":     "app",
		"{{autocomplete.name}}":      "fruit",
		"{{arg('fruit')}}":           "app",
		"{{arg('amount') * 2}}":      "6",
		"{{arg('for')}}":             "42",
		"{{arg('other') == nil}}":    "true",
		"{{interaction.command.id}}": "10",
	}
	for template, want := range tests {
		got, err := EvalTemplateToString(context.Background(), template, c)
		require.NoError(t, err, template)
		assert.Equal(t, want, got, template)
	}
}

func TestFocusedOptionWithoutFocus(t *testing.T) {
	assert.Equal(t, discord.AutocompleteOption{}, FocusedOption(nil))
}
