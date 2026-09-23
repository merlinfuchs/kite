package message

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }

var componentsV2Message = MessageData{
	Components: []ComponentData{
		{
			ID:          1,
			Type:        ComponentTypeContainer,
			AccentColor: intPtr(0xff0000),
			Components: []ComponentData{
				{
					ID:   2,
					Type: ComponentTypeSection,
					Components: []ComponentData{
						{ID: 3, Type: ComponentTypeTextDisplay, Content: "Hello {{user}}"},
					},
					Accessory: &ComponentData{ID: 4, Type: ComponentTypeButton, Style: 1, Label: "Click {{user}}", FlowSourceID: "flow-a"},
				},
				{ID: 5, Type: ComponentTypeSeparator, Divider: boolPtr(false), Spacing: 2},
				{
					ID:   6,
					Type: ComponentTypeMediaGallery,
					Items: []MediaGalleryItemData{
						{Media: UnfurledMediaItemData{URL: "https://example.com/{{user}}.png"}},
					},
				},
				{
					ID:   7,
					Type: ComponentTypeActionRow,
					Components: []ComponentData{
						{ID: 8, Type: ComponentTypeButton, Style: 2, Label: "Nested", FlowSourceID: "flow-b"},
					},
				},
			},
		},
	},
}

func TestUnmarshalLegacyActionRows(t *testing.T) {
	var data MessageData
	err := json.Unmarshal([]byte(`{"content":"hi","components":[{"id":1,"components":[{"id":2,"type":2,"style":1,"label":"A"}]}]}`), &data)
	require.NoError(t, err)

	require.Len(t, data.Components, 1)
	assert.Equal(t, ComponentTypeActionRow, data.Components[0].Type)
	assert.False(t, data.IsComponentsV2())
}

func TestEachStringReplacesNestedComponents(t *testing.T) {
	data := componentsV2Message.Copy()

	err := data.EachString(func(s *string) error {
		*s = strings.ReplaceAll(*s, "{{user}}", "merlin")
		return nil
	})
	require.NoError(t, err)

	section := data.Components[0].Components[0]
	assert.Equal(t, "Hello merlin", section.Components[0].Content)
	assert.Equal(t, "Click merlin", section.Accessory.Label)
	assert.Equal(t, "https://example.com/merlin.png", data.Components[0].Components[2].Items[0].Media.URL)

	// The original must be untouched, which also covers Copy being deep.
	assert.Equal(t, "Hello {{user}}", componentsV2Message.Components[0].Components[0].Components[0].Content)
	assert.Equal(t, "Click {{user}}", componentsV2Message.Components[0].Components[0].Accessory.Label)
}

func TestEachComponentVisitsAccessoriesAndChildren(t *testing.T) {
	var ids []int
	componentsV2Message.EachComponent(func(c *ComponentData) error {
		ids = append(ids, c.ID)
		return nil
	})

	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8}, ids)
	assert.True(t, componentsV2Message.HasInteractiveComponents())
}

func TestHasInteractiveComponentsIgnoresLinkButtons(t *testing.T) {
	data := MessageData{
		Components: []ComponentData{
			{Type: ComponentTypeActionRow, Components: []ComponentData{
				{Type: ComponentTypeButton, Style: ButtonStyleLink, URL: "https://example.com"},
			}},
		},
	}

	assert.False(t, data.HasInteractiveComponents())
}

func TestToSendMessageDataComponentsV2(t *testing.T) {
	data := componentsV2Message.Copy()
	data.Content = "should be dropped"
	data.Embeds = []EmbedData{{Description: "should be dropped"}}

	send := data.ToSendMessageData(ConvertOptions{
		ComponentIDFactory: func(c *ComponentData) discord.ComponentID {
			return discord.ComponentID(c.FlowSourceID)
		},
	})

	assert.NotZero(t, send.Flags&discord.IsComponentsV2)
	assert.Empty(t, send.Content)
	assert.Empty(t, send.Embeds)

	raw, err := json.Marshal(send.Components)
	require.NoError(t, err)

	var got []map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))
	require.Len(t, got, 1)

	container := got[0]
	assert.EqualValues(t, 17, container["type"])
	assert.EqualValues(t, 0xff0000, container["accent_color"])

	children := container["components"].([]any)
	require.Len(t, children, 4)

	section := children[0].(map[string]any)
	assert.EqualValues(t, 9, section["type"])
	accessory := section["accessory"].(map[string]any)
	assert.Equal(t, "flow-a", accessory["custom_id"])

	separator := children[1].(map[string]any)
	assert.Equal(t, false, separator["divider"])
	assert.EqualValues(t, 2, separator["spacing"])

	row := children[3].(map[string]any)
	assert.EqualValues(t, 1, row["type"])
	assert.Equal(t, "flow-b", row["components"].([]any)[0].(map[string]any)["custom_id"])
}

func TestToEditMessageDataComponentsV2ClearsContent(t *testing.T) {
	data := componentsV2Message.Copy()
	data.Content = "old"

	edit := data.ToEditMessageData(ConvertOptions{})

	raw, err := json.Marshal(edit)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))

	assert.Contains(t, got, "content")
	assert.Nil(t, got["content"])
	assert.Equal(t, []any{}, got["embeds"])
	assert.EqualValues(t, discord.IsComponentsV2, got["flags"])
}

func TestToSendMessageDataClassic(t *testing.T) {
	data := MessageData{
		Content: "hi",
		Components: []ComponentData{
			{Type: ComponentTypeActionRow, Components: []ComponentData{
				{Type: ComponentTypeButton, Style: 1, Label: "A", FlowSourceID: "flow-a"},
			}},
		},
	}

	send := data.ToSendMessageData(ConvertOptions{})

	assert.Zero(t, send.Flags&discord.IsComponentsV2)
	assert.Equal(t, "hi", send.Content)
	require.Len(t, send.Components, 1)
}

func TestToSendMessageDataStringSelect(t *testing.T) {
	data := MessageData{
		Components: []ComponentData{
			{Type: ComponentTypeActionRow, Components: []ComponentData{
				{
					ID:           7,
					Type:         ComponentTypeStringSelect,
					Placeholder:  "Pick one",
					MinValues:    1,
					MaxValues:    2,
					FlowSourceID: "flow-select",
					Options: []ComponentSelectOptionData{
						{Label: "Red", Value: "red"},
						{Label: "Blue", Value: "blue"},
					},
				},
			}},
		},
	}

	assert.True(t, data.HasInteractiveComponents())

	send := data.ToSendMessageData(ConvertOptions{})
	raw, err := json.Marshal(send.Components)
	require.NoError(t, err)

	var got []map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))

	sel := got[0]["components"].([]any)[0].(map[string]any)
	assert.EqualValues(t, 3, sel["type"])
	assert.Equal(t, "flow-select", sel["custom_id"])
	assert.Equal(t, "Pick one", sel["placeholder"])
	assert.EqualValues(t, 1, sel["min_values"])
	assert.EqualValues(t, 2, sel["max_values"])

	options := sel["options"].([]any)
	assert.Equal(t, "red", options[0].(map[string]any)["value"])
	assert.Equal(t, "blue", options[1].(map[string]any)["value"])
}
