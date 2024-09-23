package message

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func (m *MessageData) ToSendMessageData() api.SendMessageData {
	if m == nil {
		return api.SendMessageData{}
	}

	embeds := make([]discord.Embed, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.ToEmbed()
	}

	components := make(discord.ContainerComponents, len(m.Components))
	for i, component := range m.Components {
		components[i] = component.ToComponent()
	}

	return api.SendMessageData{
		Content:    m.Content,
		Flags:      discord.MessageFlags(m.Flags),
		Embeds:     embeds,
		Components: components,
	}
}

func (m *MessageData) ToEditMessageData() api.EditMessageData {
	if m == nil {
		return api.EditMessageData{}
	}

	embeds := make([]discord.Embed, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.ToEmbed()
	}

	components := make(discord.ContainerComponents, len(m.Components))
	for i, component := range m.Components {
		components[i] = component.ToComponent()
	}

	return api.EditMessageData{
		Content:    option.NewNullableString(m.Content),
		Embeds:     &embeds,
		Components: &components,
	}
}

func (m *MessageData) ToInteractionResponseData() api.InteractionResponseData {
	if m == nil {
		return api.InteractionResponseData{}
	}

	embeds := make([]discord.Embed, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.ToEmbed()
	}

	components := make(discord.ContainerComponents, len(m.Components))
	for i, component := range m.Components {
		components[i] = component.ToComponent()
	}

	return api.InteractionResponseData{
		Content:    option.NewNullableString(m.Content),
		Flags:      discord.MessageFlags(m.Flags),
		Embeds:     &embeds,
		Components: &components,
	}
}

func (m *EmbedData) ToEmbed() discord.Embed {
	if m == nil {
		return discord.Embed{}
	}

	fields := make([]discord.EmbedField, len(m.Fields))
	for i, field := range m.Fields {
		fields[i] = field.ToEmbedField()
	}

	var timestamp discord.Timestamp
	if m.Timestamp != nil {
		timestamp = discord.NewTimestamp(*m.Timestamp)
	}

	return discord.Embed{
		Title:       m.Title,
		Description: m.Description,
		URL:         m.URL,
		Timestamp:   timestamp,
		Color:       discord.Color(m.Color),
		Footer:      m.Footer.ToEmbedFooter(),
		Image:       m.Image.ToEmbedImage(),
		Thumbnail:   m.Thumbnail.ToEmbedThumbnail(),
		Author:      m.Author.ToEmbedAuthor(),
		Fields:      fields,
	}
}

func (f *EmbedFieldData) ToEmbedField() discord.EmbedField {
	if f == nil {
		return discord.EmbedField{}
	}

	if f == nil {
		return discord.EmbedField{}
	}

	return discord.EmbedField{
		Name:   f.Name,
		Value:  f.Value,
		Inline: f.Inline,
	}
}

func (f *EmbedFooterData) ToEmbedFooter() *discord.EmbedFooter {
	if f == nil {
		return nil
	}

	return &discord.EmbedFooter{
		Text: f.Text,
		Icon: f.IconURL,
	}
}

func (i *EmbedImageData) ToEmbedImage() *discord.EmbedImage {
	if i == nil {
		return nil
	}

	return &discord.EmbedImage{
		URL: i.URL,
	}
}

func (t *EmbedThumbnailData) ToEmbedThumbnail() *discord.EmbedThumbnail {
	if t == nil {
		return nil
	}

	return &discord.EmbedThumbnail{
		URL: t.URL,
	}
}

func (a *EmbedAuthorData) ToEmbedAuthor() *discord.EmbedAuthor {
	if a == nil {
		return nil
	}

	return &discord.EmbedAuthor{
		Name: a.Name,
		URL:  a.URL,
		Icon: a.IconURL,
	}
}

func (r *ComponentRowData) ToComponent() discord.ContainerComponent {
	if r == nil {
		return nil
	}

	components := make(discord.ActionRowComponent, len(r.Components))
	for i, component := range r.Components {
		components[i] = component.ToComponent()
	}

	return &components
}

func (c *ComponentData) ToComponent() discord.InteractiveComponent {
	if c == nil {
		return nil
	}

	switch c.Type {
	case int(discord.ButtonComponentType):
		var style discord.ButtonComponentStyle
		switch c.Style {
		case 2:
			style = discord.SecondaryButtonStyle()
		case 3:
			style = discord.SuccessButtonStyle()
		case 4:
			style = discord.DangerButtonStyle()
		case 5:
			style = discord.LinkButtonStyle(c.URL)
		default:
			style = discord.PrimaryButtonStyle()
		}

		var customID discord.ComponentID
		if c.Style != 5 {
			customID = discord.ComponentID(c.FlowSourceID)
		}

		return &discord.ButtonComponent{
			Style:    style,
			Label:    c.Label,
			Emoji:    c.Emoji.ToEmoji(),
			Disabled: c.Disabled,
			CustomID: customID,
		}
	}

	return nil
}

func (e *ComponentEmojiData) ToEmoji() *discord.ComponentEmoji {
	if e == nil {
		return nil
	}

	id, _ := discord.ParseSnowflake(e.ID)

	return &discord.ComponentEmoji{
		Name:     e.Name,
		ID:       discord.EmojiID(id),
		Animated: e.Animated,
	}
}
