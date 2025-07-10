package message

func (m *MessageData) Copy() MessageData {
	if m == nil {
		return MessageData{}
	}

	embeds := make([]EmbedData, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.Copy()
	}

	components := make([]ComponentRowData, len(m.Components))
	for i, component := range m.Components {
		components[i] = component.Copy()
	}

	attachments := make([]MessageAttachment, len(m.Attachments))
	for i, attachment := range m.Attachments {
		attachments[i] = attachment.Copy()
	}

	return MessageData{
		Content:         m.Content,
		Flags:           m.Flags,
		Embeds:          embeds,
		Attachments:     attachments,
		Components:      components,
		AllowedMentions: m.AllowedMentions.Copy(),
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

func (f *EmbedFooterData) Copy() *EmbedFooterData {
	if f == nil {
		return nil
	}

	return &EmbedFooterData{
		Text:    f.Text,
		IconURL: f.IconURL,
	}
}

func (f *EmbedImageData) Copy() *EmbedImageData {
	if f == nil {
		return nil
	}

	return &EmbedImageData{
		URL: f.URL,
	}
}

func (f *EmbedThumbnailData) Copy() *EmbedThumbnailData {
	if f == nil {
		return nil
	}

	return &EmbedThumbnailData{
		URL: f.URL,
	}
}

func (a *EmbedAuthorData) Copy() *EmbedAuthorData {
	if a == nil {
		return nil
	}

	return &EmbedAuthorData{
		Name:    a.Name,
		URL:     a.URL,
		IconURL: a.IconURL,
	}
}

func (c ComponentRowData) Copy() ComponentRowData {
	components := make([]ComponentData, len(c.Components))
	for i, component := range c.Components {
		components[i] = component.Copy()
	}

	return ComponentRowData{
		ID:         c.ID,
		Components: components,
	}
}

func (c ComponentData) Copy() ComponentData {
	options := make([]ComponentSelectOptionData, len(c.Options))
	for i, option := range c.Options {
		options[i] = option.Copy()
	}

	return ComponentData{
		ID:           c.ID,
		Type:         c.Type,
		Disabled:     c.Disabled,
		Style:        c.Style,
		Label:        c.Label,
		Emoji:        c.Emoji.Copy(),
		URL:          c.URL,
		Placeholder:  c.Placeholder,
		MinValues:    c.MinValues,
		MaxValues:    c.MaxValues,
		Options:      options,
		FlowSourceID: c.FlowSourceID,
	}
}

func (c *ComponentEmojiData) Copy() *ComponentEmojiData {
	if c == nil {
		return nil
	}

	return &ComponentEmojiData{
		Name:     c.Name,
		ID:       c.ID,
		Animated: c.Animated,
	}
}

func (c ComponentSelectOptionData) Copy() ComponentSelectOptionData {
	return ComponentSelectOptionData{
		ID:           c.ID,
		Label:        c.Label,
		Description:  c.Description,
		Emoji:        c.Emoji.Copy(),
		Default:      c.Default,
		FlowSourceID: c.FlowSourceID,
	}
}

func (c MessageAttachment) Copy() MessageAttachment {
	return MessageAttachment{
		AssetID: c.AssetID,
	}
}

func (a *AllowedMentionsData) Copy() *AllowedMentionsData {
	if a == nil {
		return nil
	}

	parse := make([]string, len(a.Parse))
	copy(parse, a.Parse)

	return &AllowedMentionsData{
		Parse: parse,
	}
}
