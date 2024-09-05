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

	return api.SendMessageData{
		Content: m.Content,
		Flags:   discord.MessageFlags(m.Flags),
		Embeds:  embeds,
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

	return api.InteractionResponseData{
		Content: option.NewNullableString(m.Content),
		Flags:   discord.MessageFlags(m.Flags),
		Embeds:  &embeds,
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

	return discord.Embed{
		Title:       m.Title,
		Description: m.Description,
		URL:         m.URL,
		Timestamp:   discord.NewTimestamp(m.Timestamp),
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
