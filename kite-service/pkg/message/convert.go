package message

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

type ConvertOptions struct {
	ComponentIDFactory componentIDFactory
}

func (m *MessageData) ToSendMessageData(opts ConvertOptions) api.SendMessageData {
	if m == nil {
		return api.SendMessageData{}
	}

	embeds := make([]discord.Embed, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.ToEmbed()
	}

	components := make(discord.ContainerComponents, len(m.Components))
	for i, component := range m.Components {
		components[i] = component.ToComponent(opts)
	}

	return api.SendMessageData{
		Content:         m.Content,
		Flags:           discord.MessageFlags(m.Flags),
		Embeds:          embeds,
		Components:      components,
		AllowedMentions: m.AllowedMentions.ToAllowedMentions(),
	}
}

func (m *MessageData) ToEditMessageData(opts ConvertOptions) api.EditMessageData {
	if m == nil {
		return api.EditMessageData{}
	}

	embeds := make([]discord.Embed, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.ToEmbed()
	}

	components := make(discord.ContainerComponents, len(m.Components))
	for i, component := range m.Components {
		components[i] = component.ToComponent(opts)
	}

	var flags *discord.MessageFlags
	if m.Flags != 0 {
		f := discord.MessageFlags(m.Flags)
		flags = &f
	}

	return api.EditMessageData{
		Content:         option.NewNullableString(m.Content),
		Flags:           flags,
		Embeds:          &embeds,
		Components:      &components,
		AllowedMentions: m.AllowedMentions.ToAllowedMentions(),
	}
}

func (m *MessageData) ToInteractionResponseData(opts ConvertOptions) api.InteractionResponseData {
	if m == nil {
		return api.InteractionResponseData{}
	}

	embeds := make([]discord.Embed, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.ToEmbed()
	}

	components := make(discord.ContainerComponents, len(m.Components))
	for i, component := range m.Components {
		components[i] = component.ToComponent(opts)
	}

	return api.InteractionResponseData{
		Content:         option.NewNullableString(m.Content),
		Flags:           discord.MessageFlags(m.Flags),
		Embeds:          &embeds,
		Components:      &components,
		AllowedMentions: m.AllowedMentions.ToAllowedMentions(),
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

func (r *ComponentRowData) ToComponent(opts ConvertOptions) discord.ContainerComponent {
	if r == nil {
		return nil
	}

	components := make(discord.ActionRowComponent, len(r.Components))
	for i, component := range r.Components {
		components[i] = component.ToComponent(opts)
	}

	return &components
}

func (c *ComponentData) ToComponent(opts ConvertOptions) discord.InteractiveComponent {
    if c == nil {
        return nil
    }

    switch c.Type {
    case 1: // ActionRow
        return c.toActionRow(opts)
    case 2: // Button
        return c.toButton(opts)
    case 3: // StringSelect
        return c.toSelectMenu(opts)
    case 9: // Section
        return c.toSection(opts)
    case 10: // TextDisplay
        return c.toTextDisplay(opts)
    case 11: // Thumbnail
        return c.toThumbnail(opts)
    case 12: // MediaGallery
        return c.toMediaGallery(opts)
    case 13: // File
        return c.toFile(opts)
    case 14: // Separator
        return c.toSeparator(opts)
    case 17: // Container
        return c.toContainer(opts)
    default:
        return nil
    }
}

// button logic
func (c *ComponentData) toButton(opts ConvertOptions) *discord.ButtonComponent {
    var customID discord.ComponentID
    if opts.ComponentIDFactory != nil {
        customID = opts.ComponentIDFactory(c)
    } else {
        customID = discord.ComponentID(c.FlowSourceID)
    }

    button := &discord.ButtonComponent{
        CustomID: customID,
        Label:    c.Label,
        Disabled: c.Disabled,
    }

    if c.Emoji != nil {
        button.Emoji = c.Emoji.ToEmoji()
    }

    if c.Style >= 1 && c.Style <= 5 {
        button.Style = discord.ButtonStyle(c.Style)
    }

    if c.URL != "" {
        button.URL = c.URL
        button.Style = discord.LinkButtonStyle
    }

    return button
}

// select menu logic
func (c *ComponentData) toSelectMenu(opts ConvertOptions) *discord.StringSelectComponent {
    var customID discord.ComponentID
    if opts.ComponentIDFactory != nil {
        customID = opts.ComponentIDFactory(c)
    } else {
        customID = discord.ComponentID(c.FlowSourceID)
    }

    options := make([]discord.SelectOption, len(c.Options))
    for i, opt := range c.Options {
        options[i] = opt.ToSelectOption()
    }

    return &discord.StringSelectComponent{
        CustomID:    customID,
        Placeholder: c.Placeholder,
        Disabled:    c.Disabled,
        Options:     options,
        ValueLimits: [2]int{c.MinValues, c.MaxValues},
    }
}

// ActionRow with recursive children
func (c *ComponentData) toActionRow(opts ConvertOptions) *discord.ActionRowComponent {
    components := make([]discord.InteractiveComponent, 0, len(c.Components))
    
    for _, child := range c.Components {
        if comp := child.ToComponent(opts); comp != nil {
            components = append(components, comp)
        }
    }
    
    return &discord.ActionRowComponent{
        Components: components,
    }
}

// Section component
func (c *ComponentData) toSection(opts ConvertOptions) *discord.SectionComponent {
    section := &discord.SectionComponent{
        Components: make([]discord.InteractiveComponent, 0, len(c.Components)),
    }
    
    for _, child := range c.Components {
        if comp := child.ToComponent(opts); comp != nil {
            section.Components = append(section.Components, comp)
        }
    }
    
    if c.Accessory != nil {
        section.Accessory = c.Accessory.ToComponent(opts)
    }
    
    return section
}

// Text Display
func (c *ComponentData) toTextDisplay(opts ConvertOptions) *discord.TextDisplayComponent {
    return &discord.TextDisplayComponent{
        Content: c.Content,
    }
}

// Thumbnail
func (c *ComponentData) toThumbnail(opts ConvertOptions) *discord.ThumbnailComponent {
    if c.Media == nil {
        return nil
    }
    
    thumb := &discord.ThumbnailComponent{
        Media:       discord.UnfurledMediaItem{URL: c.Media.URL},
        Description: c.Description,
        Spoiler:     c.Spoiler,
    }
    
    if c.Content != "" {
        thumb.Content = c.Content
    }
    
    return thumb
}

// Media Gallery
func (c *ComponentData) toMediaGallery(opts ConvertOptions) *discord.MediaGalleryComponent {
    items := make([]discord.MediaGalleryItem, len(c.Items))
    
    for i, item := range c.Items {
        items[i] = discord.MediaGalleryItem{
            Media:       discord.UnfurledMediaItem{URL: item.Media.URL},
            Description: item.Description,
            Spoiler:     item.Spoiler,
        }
    }
    
    return &discord.MediaGalleryComponent{
        Items: items,
    }
}

// File
func (c *ComponentData) toFile(opts ConvertOptions) *discord.FileComponent {
    if c.File == nil {
        return nil
    }
    
    return &discord.FileComponent{
        File:    discord.UnfurledMediaItem{URL: c.File.URL},
        Spoiler: c.Spoiler,
    }
}

// Separator
func (c *ComponentData) toSeparator(opts ConvertOptions) *discord.SeparatorComponent {
    return &discord.SeparatorComponent{
        Divider: c.Divider,
        Spacing: c.Spacing,
    }
}

// Container
func (c *ComponentData) toContainer(opts ConvertOptions) *discord.ContainerComponent {
    container := &discord.ContainerComponent{
        Components: make([]discord.InteractiveComponent, 0, len(c.Components)),
        Spoiler:    c.Spoiler,
    }
    
    for _, child := range c.Components {
        if comp := child.ToComponent(opts); comp != nil {
            container.Components = append(container.Components, comp)
        }
    }
    
    if c.AccentColor != nil {
        container.AccentColor = *c.AccentColor
    }
    
    return container
}

func (o *ComponentSelectOptionData) ToSelectOption() discord.SelectOption {
	if o == nil {
		return discord.SelectOption{}
	}

	var emoji *discord.ComponentEmoji
	if o.Emoji != nil {
		emoji = o.Emoji.ToEmoji()
	}

	return discord.SelectOption{
		Label:       o.Label,
		Value:       o.FlowSourceID,
		Description: o.Description,
		Emoji:       emoji,
		Default:     o.Default,
	}
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

func (a *AllowedMentionsData) ToAllowedMentions() *api.AllowedMentions {
	if a == nil {
		return &api.AllowedMentions{
			Parse: []api.AllowedMentionType{
				api.AllowUserMention,
			},
		}
	}

	parse := make([]api.AllowedMentionType, len(a.Parse))
	for i, p := range a.Parse {
		parse[i] = api.AllowedMentionType(p)
	}

	return &api.AllowedMentions{
		Parse: parse,
	}
}

type componentIDFactory func(component *ComponentData) discord.ComponentID
