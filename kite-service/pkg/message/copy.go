package message

func (m *MessageData) Copy() MessageData {
	if m == nil {
		return MessageData{}
	}

	embeds := make([]EmbedData, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.Copy()
	}

	return MessageData{
		Content: m.Content,
		Embeds:  embeds,
	}
}

func (m EmbedData) Copy() EmbedData {
	fields := make([]EmbedFieldData, len(m.Fields))
	for i, field := range m.Fields {
		fields[i] = field.Copy()
	}

	return EmbedData{
		Title:       m.Title,
		Description: m.Description,
		URL:         m.URL,
		Timestamp:   m.Timestamp,
		Color:       m.Color,
		Author:      m.Author.Copy(),
		Footer:      m.Footer.Copy(),
		Image:       m.Image.Copy(),
		Thumbnail:   m.Thumbnail.Copy(),
	}
}

func (f EmbedFieldData) Copy() EmbedFieldData {
	return EmbedFieldData{
		Name:   f.Name,
		Value:  f.Value,
		Inline: f.Inline,
	}
}

func (f EmbedFooterData) Copy() *EmbedFooterData {
	return &EmbedFooterData{
		Text:    f.Text,
		IconURL: f.IconURL,
	}
}

func (f EmbedImageData) Copy() *EmbedImageData {
	return &EmbedImageData{
		URL: f.URL,
	}
}

func (f EmbedThumbnailData) Copy() *EmbedThumbnailData {
	return &EmbedThumbnailData{
		URL: f.URL,
	}
}

func (a EmbedAuthorData) Copy() *EmbedAuthorData {
	return &EmbedAuthorData{
		Name:    a.Name,
		URL:     a.URL,
		IconURL: a.IconURL,
	}
}
