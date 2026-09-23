package message

func (m *MessageData) Copy() MessageData {
	if m == nil {
		return MessageData{}
	}

	embeds := make([]EmbedData, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.Copy()
	}

	components := copyComponents(m.Components)

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
		Fields:      fields,
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

func copyComponents(components []ComponentData) []ComponentData {
	if components == nil {
		return nil
	}

	res := make([]ComponentData, len(components))
	for i, component := range components {
		res[i] = component.Copy()
	}
	return res
}

func (c ComponentData) Copy() ComponentData {
	res := c
	res.Emoji = c.Emoji.Copy()
	res.Components = copyComponents(c.Components)

	if c.Options != nil {
		res.Options = make([]ComponentSelectOptionData, len(c.Options))
		for i, option := range c.Options {
			res.Options[i] = option.Copy()
		}
	}

	if c.Accessory != nil {
		accessory := c.Accessory.Copy()
		res.Accessory = &accessory
	}

	if c.Media != nil {
		media := *c.Media
		res.Media = &media
	}

	if c.File != nil {
		file := *c.File
		res.File = &file
	}

	if c.Items != nil {
		res.Items = make([]MediaGalleryItemData, len(c.Items))
		copy(res.Items, c.Items)
	}

	if c.Divider != nil {
		divider := *c.Divider
		res.Divider = &divider
	}

	if c.AccentColor != nil {
		accentColor := *c.AccentColor
		res.AccentColor = &accentColor
	}

	return res
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
		Value:        c.Value,
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
