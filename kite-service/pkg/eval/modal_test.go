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
	assert.Equal(t, []string{"legacy"}, components["old"].Values)
	assert.Equal(t, "Jack", components["name"].Value)
	assert.Equal(t, []string{"Jack"}, components["name"].Values)
	assert.Equal(t, "red", components["colors"].Value)
	assert.Equal(t, []string{"red", "blue"}, components["colors"].Values)
	assert.Equal(t, "123", components["users"].Value)
	assert.Equal(t, []string{"123"}, components["users"].Values)
	assert.Equal(t, "large", components["size"].Value)
	assert.Equal(t, []string{"large"}, components["size"].Values)
	assert.Equal(t, "", components["extras"].Value)
	assert.Empty(t, components["extras"].Values)
	assert.Equal(t, "true", components["agree"].Value)
	assert.Equal(t, []string{"true"}, components["agree"].Values)
}

func TestModalInputValues(t *testing.T) {
	data := `{
		"custom_id": "modal",
		"components": [
			{"type": 18, "id": 1, "component": {"type": 3, "custom_id": "color", "values": ["red"]}},
			{"type": 18, "id": 2, "component": {"type": 22, "custom_id": "extras", "values": ["cheese", "bacon"]}},
			{"type": 18, "id": 3, "component": {"type": 22, "custom_id": "none", "values": []}},
			{"type": 18, "id": 4, "component": {"type": 3, "custom_id": "colors", "values": ["red", "blue"]}},
			{"type": 18, "id": 5, "component": {"type": 6, "custom_id": "roles", "values": ["1", "2"]}},
			{"type": 18, "id": 6, "component": {"type": 4, "custom_id": "name", "value": "Jack"}},
			{"type": 18, "id": 7, "component": {"type": 4, "custom_id": "empty", "value": ""}},
			{"type": 18, "id": 8, "component": {"type": 21, "custom_id": "size", "value": "large"}},
			{"type": 18, "id": 9, "component": {"type": 21, "custom_id": "nosize"}},
			{"type": 18, "id": 10, "component": {"type": 23, "custom_id": "agree", "value": true}},
			{"type": 18, "id": 11, "component": {"type": 23, "custom_id": "decline", "value": false}}
		]
	}`

	var modal discord.ModalInteraction
	require.NoError(t, json.Unmarshal([]byte(data), &modal))

	c := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &modal,
	}, state.New("Bot test"))

	tests := map[string]string{
		// input() returns the first pick, like interaction.value.
		"{{input('color')}}":                                     "red",
		"{{input('extras')}}":                                    "cheese",
		"Extras: {{input('extras')}}":                            "Extras: cheese",
		"{{input('none')}}":                                      "",
		"{{input('roles')}}":                                     "1",
		"{{input('name')}}":                                      "Jack",
		"{{input('size')}}":                                      "large",
		"{{input('nosize')}}":                                    "",
		"{{input('agree')}}":                                     "true",
		"{{input('decline')}}":                                   "false",
		"{{input('unknown') == nil}}":                            "true",
		"{{interaction.components['extras'].value}}":             "cheese",
		"{{interaction.components['extras'].values[1]}}":         "bacon",
		"{{len(interaction.components['extras'].values)}}":       "2",
		"{{'bacon' in interaction.components['extras'].values}}": "true",
		"{{'bac' in interaction.components['extras'].values}}":   "false",

		// inputs() returns every pick, or the one value of other inputs.
		"{{len(inputs('color'))}}":         "1",
		"{{inputs('extras')[1]}}":          "bacon",
		"{{len(inputs('none'))}}":          "0",
		"{{join(inputs('roles'), ',')}}":   "1,2",
		"{{join(inputs('name'), ',')}}":    "Jack",
		"{{len(inputs('empty'))}}":         "0",
		"{{join(inputs('size'), ',')}}":    "large",
		"{{len(inputs('nosize'))}}":        "0",
		"{{join(inputs('agree'), ',')}}":   "true",
		"{{join(inputs('decline'), ',')}}": "false",
		"{{inputs('unknown') == nil}}":     "true",

		// The examples of the docs.
		"{{'red' in inputs('colors')}}":      "true",
		"{{len(inputs('colors'))}}":          "2",
		"{{join(inputs('colors'), ', ')}}":   "red, blue",
		"{{ 'red' in inputs('colors') }}":    "true",
		"{{ len(inputs('colors')) }}":        "2",
		"{{ join(inputs('colors'), ', ') }}": "red, blue",
	}
	for template, want := range tests {
		got, err := EvalTemplateToString(context.Background(), template, c)
		require.NoError(t, err, template)
		assert.Equal(t, want, got, template)
	}
}

// The inputs of a modal before a resume point still work after it.
func TestModalInputsAfterResume(t *testing.T) {
	data := `{
		"custom_id": "modal",
		"components": [
			{"type": 18, "id": 1, "component": {"type": 22, "custom_id": "extras", "values": ["cheese", "bacon"]}}
		]
	}`

	var modal discord.ModalInteraction
	require.NoError(t, json.Unmarshal([]byte(data), &modal))

	session := state.New("Bot test")
	submit := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &modal,
	}, session)
	click := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &discord.ButtonInteraction{CustomID: "button"},
	}, session)
	click.SetResumeContext([]Context{submit})

	tests := map[string]string{
		"{{input('extras')}}":              "cheese",
		"{{join(inputs('extras'), ', ')}}": "cheese, bacon",
		"{{'bacon' in inputs('extras')}}":  "true",
		"{{inputs('unknown') == nil}}":     "true",
		"{{len(origin.inputs('extras'))}}": "2",
	}
	for template, want := range tests {
		got, err := EvalTemplateToString(context.Background(), template, click)
		require.NoError(t, err, template)
		assert.Equal(t, want, got, template)
	}
}

func TestModalPicked(t *testing.T) {
	data := `{
		"custom_id": "modal",
		"components": [
			{"type": 18, "id": 1, "component": {"type": 5, "custom_id": "users", "values": ["10", "11"]}},
			{"type": 18, "id": 2, "component": {"type": 6, "custom_id": "roles", "values": ["20"]}},
			{"type": 18, "id": 3, "component": {"type": 7, "custom_id": "mentions", "values": ["20", "10"]}},
			{"type": 18, "id": 4, "component": {"type": 8, "custom_id": "channels", "values": ["30"]}},
			{"type": 18, "id": 5, "component": {"type": 5, "custom_id": "nobody", "values": []}},
			{"type": 18, "id": 6, "component": {"type": 4, "custom_id": "name", "value": "Jack"}}
		],
		"resolved": {
			"users": {
				"10": {"id": "10", "username": "jack"},
				"11": {"id": "11", "username": "jill", "global_name": "Jill"}
			},
			"members": {"10": {"nick": "Jacky", "roles": ["20"]}},
			"roles": {"20": {"id": "20", "name": "Mods"}},
			"channels": {"30": {"id": "30", "name": "support", "type": 0}}
		}
	}`

	var modal discord.ModalInteraction
	require.NoError(t, json.Unmarshal([]byte(data), &modal))

	c := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &modal,
	}, state.New("Bot test"))

	tests := map[string]string{
		// Like arg(), they print as mentions.
		"{{picked('users')}}":                                 "<@10>",
		"{{picked('users').id}}":                              "10",
		"{{picked('channels')}}":                              "<#30>",
		"{{picked('users').username}}":                        "jack",
		"{{picked('users').nick}}":                            "Jacky",
		"{{pickedAll('users')[1].display_name}}":              "Jill",
		"{{len(pickedAll('users'))}}":                         "2",
		"{{join(map(pickedAll('users'), #.username), ', ')}}": "jack, jill",
		"{{picked('roles').name}}":                            "Mods",
		"{{picked('mentions').name}}":                         "Mods",
		"{{pickedAll('mentions')[1].username}}":               "jack",
		"{{picked('channels').name}}":                         "support",
		"{{picked('nobody')}}":                                "",
		"{{picked('nobody').username}}":                       "",
		"{{len(pickedAll('nobody'))}}":                        "0",
		"{{picked('name') == nil}}":                           "true",
		"{{picked('unknown') == nil}}":                        "true",
		"{{pickedAll('unknown') == nil}}":                     "true",

		// The examples of the docs.
		"{{picked('users').display_name}}":                "jack",
		"{{join(map(pickedAll('roles'), #.name), ', ')}}": "Mods",

		// input() still returns the IDs.
		"{{input('users') == '10'}}":     "true",
		"{{'20' in inputs('mentions')}}": "true",
	}
	for template, want := range tests {
		got, err := EvalTemplateToString(context.Background(), template, c)
		require.NoError(t, err, template)
		assert.Equal(t, want, got, template)
	}
}

// The picks of a modal before a resume point still work after it, also once
// the interaction was stored and loaded again.
func TestModalPickedAfterResume(t *testing.T) {
	data := `{
		"custom_id": "modal",
		"components": [
			{"type": 18, "id": 1, "component": {"type": 6, "custom_id": "roles", "values": ["20"]}}
		],
		"resolved": {"roles": {"20": {"id": "20", "name": "Mods"}}}
	}`

	stored, err := json.Marshal(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &discord.ModalInteraction{},
	})
	require.NoError(t, err)
	var submitEvent discord.InteractionEvent
	require.NoError(t, json.Unmarshal(stored, &submitEvent))
	modal := submitEvent.Data.(*discord.ModalInteraction)
	require.NoError(t, json.Unmarshal([]byte(data), modal))

	stored, err = json.Marshal(&submitEvent)
	require.NoError(t, err)
	var loaded discord.InteractionEvent
	require.NoError(t, json.Unmarshal(stored, &loaded))

	session := state.New("Bot test")
	submit := NewContextFromInteraction(&loaded, session)
	click := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &discord.ButtonInteraction{CustomID: "button"},
	}, session)
	click.SetResumeContext([]Context{submit})

	tests := map[string]string{
		"{{picked('roles').name}}":        "Mods",
		"{{pickedAll('roles')[0].name}}":  "Mods",
		"{{origin.picked('roles').name}}": "Mods",
		"{{picked('unknown') == nil}}":    "true",
	}
	for template, want := range tests {
		got, err := EvalTemplateToString(context.Background(), template, click)
		require.NoError(t, err, template)
		assert.Equal(t, want, got, template)
	}
}

// A file upload's answers are the IDs of the uploaded files, and picked()
// returns the files, which Discord sends as resolved attachments.
func TestModalFileUpload(t *testing.T) {
	data := `{
		"custom_id": "modal",
		"components": [
			{"type": 18, "id": 1, "component": {"type": 19, "custom_id": "image", "values": ["111"]}},
			{"type": 18, "id": 2, "component": {"type": 19, "custom_id": "images", "values": ["222", "333"]}},
			{"type": 18, "id": 3, "component": {"type": 19, "custom_id": "none", "values": []}}
		],
		"resolved": {
			"attachments": {
				"111": {"id": "111", "filename": "cat.png", "content_type": "image/png", "size": 1024, "url": "https://cdn.discordapp.com/cat.png", "proxy_url": "https://media.discordapp.net/cat.png"},
				"222": {"id": "222", "filename": "a.png", "size": 1, "url": "https://cdn.discordapp.com/a.png", "proxy_url": "https://media.discordapp.net/a.png"},
				"333": {"id": "333", "filename": "b.pdf", "size": 1, "url": "https://cdn.discordapp.com/b.pdf", "proxy_url": "https://media.discordapp.net/b.pdf"}
			}
		}
	}`

	var modal discord.ModalInteraction
	require.NoError(t, json.Unmarshal([]byte(data), &modal))

	c := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &modal,
	}, state.New("Bot test"))

	tests := map[string]string{
		"{{input('image')}}":                                  "111",
		"{{join(inputs('images'), ' ')}}":                     "222 333",
		"{{interaction.components['images'].values[1]}}":      "333",
		"{{input('none')}}":                                   "",
		"{{len(inputs('none'))}}":                             "0",
		"{{picked('image').filename}}":                        "cat.png",
		"{{picked('image').content_type}}":                    "image/png",
		"{{picked('image').size}}":                            "1024",
		"{{picked('image').proxy_url}}":                       "https://media.discordapp.net/cat.png",
		"{{len(pickedAll('images'))}}":                        "2",
		"{{join(map(pickedAll('images'), #.filename), ' ')}}": "a.png b.pdf",
		"{{picked('none')}}":                                  "",
		"{{picked('none')?.url}}":                             "",
		"{{len(pickedAll('none'))}}":                          "0",

		// Like an attachment argument, a file prints as its URL.
		"{{picked('image')}}": "https://cdn.discordapp.com/cat.png",
	}
	for template, want := range tests {
		got, err := EvalTemplateToString(context.Background(), template, c)
		require.NoError(t, err, template)
		assert.Equal(t, want, got, template)
	}
}

// The files of a modal before a resume point still work after it, as the
// resolved attachments are stored with the interaction.
func TestModalFileUploadAfterResume(t *testing.T) {
	data := `{
		"custom_id": "modal",
		"components": [
			{"type": 18, "id": 1, "component": {"type": 19, "custom_id": "image", "values": ["111"]}}
		],
		"resolved": {
			"attachments": {
				"111": {"id": "111", "filename": "cat.png", "size": 1, "url": "https://cdn.discordapp.com/cat.png", "proxy_url": "https://media.discordapp.net/cat.png"}
			}
		}
	}`

	var modal discord.ModalInteraction
	require.NoError(t, json.Unmarshal([]byte(data), &modal))

	// The interaction is stored as JSON at a resume point.
	b, err := json.Marshal(&discord.InteractionEvent{User: &discord.User{ID: 1}, Data: &modal})
	require.NoError(t, err)
	var stored discord.InteractionEvent
	require.NoError(t, json.Unmarshal(b, &stored))

	session := state.New("Bot test")
	submit := NewContextFromInteraction(&stored, session)
	click := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &discord.ButtonInteraction{CustomID: "button"},
	}, session)
	click.SetResumeContext([]Context{submit})

	tests := map[string]string{
		"{{input('image')}}":  "111",
		"{{picked('image')}}": "https://cdn.discordapp.com/cat.png",
	}
	for template, want := range tests {
		got, err := EvalTemplateToString(context.Background(), template, click)
		require.NoError(t, err, template)
		assert.Equal(t, want, got, template)
	}
}
