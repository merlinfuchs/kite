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

	data := api.SendMessageData{
		Content:         m.Content,
		Flags:           m.messageFlags(),
		Embeds:          m.toEmbeds(),
		Components:      m.toComponents(opts),
		AllowedMentions: m.AllowedMentions.ToAllowedMentions(),
	}

	if m.IsComponentsV2() {
		data.Content = ""
		data.Embeds = nil
	}

	return data
}

func (m *MessageData) ToEditMessageData(opts ConvertOptions) api.EditMessageData {
	if m == nil {
		return api.EditMessageData{}
	}

	embeds := m.toEmbeds()
	components := m.toComponents(opts)

	var flags *discord.MessageFlags
	if f := m.messageFlags(); f != 0 {
		flags = &f
	}

	data := api.EditMessageData{
		Content:         option.NewNullableString(m.Content),
		Flags:           flags,
		Embeds:          &embeds,
		Components:      &components,
		AllowedMentions: m.AllowedMentions.ToAllowedMentions(),
	}

	if m.IsComponentsV2() {
		// Content and embeds have to be cleared when a classic message is turned into a components v2 message.
		data.Content = option.NullString
		data.Embeds = &[]discord.Embed{}
	}

	return data
}

func (m *MessageData) ToInteractionResponseData(opts ConvertOptions) api.InteractionResponseData {
	if m == nil {
		return api.InteractionResponseData{}
	}

	embeds := m.toEmbeds()
	components := m.toComponents(opts)

	data := api.InteractionResponseData{
		Content:         option.NewNullableString(m.Content),
		Flags:           m.messageFlags(),
		Embeds:          &embeds,
		Components:      &components,
		AllowedMentions: m.AllowedMentions.ToAllowedMentions(),
	}

	if m.IsComponentsV2() {
		data.Content = option.NullString
		data.Embeds = &[]discord.Embed{}
	}

	return data
}

func (m *MessageData) messageFlags() discord.MessageFlags {
	flags := discord.MessageFlags(m.Flags)
	if m.IsComponentsV2() {
		flags |= discord.IsComponentsV2
	}
	return flags
}

func (m *MessageData) toEmbeds() []discord.Embed {
	embeds := make([]discord.Embed, len(m.Embeds))
	for i, embed := range m.Embeds {
		embeds[i] = embed.ToEmbed()
	}
	return embeds
}

func (m *MessageData) toComponents(opts ConvertOptions) discord.TopLevelComponents {
	components := make(discord.TopLevelComponents, 0, len(m.Components))
	for i := range m.Components {
		if c, ok := m.Components[i].ToComponent(opts).(discord.TopLevelComponent); ok {
			components = append(components, c)
		}
	}
	return components
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

// ToComponent converts the component to its arikawa representation, returning nil for unsupported types.
func (c *ComponentData) ToComponent(opts ConvertOptions) discord.Component {
	if c == nil {
		return nil
	}

	switch c.Type {
	case ComponentTypeActionRow:
		row := make(discord.ActionRowComponent, 0, len(c.Components))
		for i := range c.Components {
			if ic, ok := c.Components[i].ToComponent(opts).(discord.InteractiveComponent); ok {
				row = append(row, ic)
			}
		}
		return &row
	case ComponentTypeButton:
		return c.toButton(opts)
	case ComponentTypeStringSelect:
		return c.toStringSelect(opts)
	case ComponentTypeSection:
		return &discord.SectionComponent{
			Components: c.childComponents(opts),
			Accessory:  c.Accessory.ToComponent(opts),
		}
	case ComponentTypeTextDisplay:
		return &discord.TextDisplayComponent{
			Content: c.Content,
		}
	case ComponentTypeThumbnail:
		return &discord.ThumbnailComponent{
			Media:       c.Media.toUnfurledMediaItem(),
			Description: c.Description,
			Spoiler:     c.Spoiler,
		}
	case ComponentTypeMediaGallery:
		items := make([]discord.MediaGalleryComponentItem, len(c.Items))
		for i, item := range c.Items {
			items[i] = discord.MediaGalleryComponentItem{
				Media:       item.Media.toUnfurledMediaItem(),
				Description: item.Description,
				Spoiler:     item.Spoiler,
			}
		}
		return &discord.MediaGalleryComponent{
			Items: items,
		}
	case ComponentTypeFile:
		return &discord.FileComponent{
			File:    c.File.toUnfurledMediaItem(),
			Spoiler: c.Spoiler,
		}
	case ComponentTypeSeparator:
		return &discord.SeparatorComponent{
			Divider: option.Bool(c.Divider),
			Spacing: discord.SeparatorComponentSpacing(c.Spacing),
		}
	case ComponentTypeContainer:
		var accentColor discord.Color
		if c.AccentColor != nil {
			accentColor = discord.Color(*c.AccentColor)
		}
		return &discord.ContainerComponent{
			Components:  c.childComponents(opts),
			AccentColor: accentColor,
			Spoiler:     c.Spoiler,
		}
	}

	return nil
}

func (c *ComponentData) childComponents(opts ConvertOptions) []discord.Component {
	components := make([]discord.Component, 0, len(c.Components))
	for i := range c.Components {
		if child := c.Components[i].ToComponent(opts); child != nil {
			components = append(components, child)
		}
	}
	return components
}

func (c *ComponentData) toButton(opts ConvertOptions) *discord.ButtonComponent {
	var style discord.ButtonComponentStyle
	switch c.Style {
	case 2:
		style = discord.SecondaryButtonStyle()
	case 3:
		style = discord.SuccessButtonStyle()
	case 4:
		style = discord.DangerButtonStyle()
	case ButtonStyleLink:
		style = discord.LinkButtonStyle(c.URL)
	default:
		style = discord.PrimaryButtonStyle()
	}

	var customID discord.ComponentID
	if c.Style != ButtonStyleLink {
		customID = c.customID(opts)
	}

	return &discord.ButtonComponent{
		Style:    style,
		Label:    c.Label,
		Emoji:    c.Emoji.ToEmoji(),
		Disabled: c.Disabled,
		CustomID: customID,
	}
}

func (c *ComponentData) toStringSelect(opts ConvertOptions) *discord.StringSelectComponent {
	options := make([]discord.SelectOption, len(c.Options))
	for i, option := range c.Options {
		value := option.Value
		if value == "" {
			value = option.Label
		}

		options[i] = discord.SelectOption{
			Label:       option.Label,
			Value:       value,
			Description: option.Description,
			Emoji:       option.Emoji.ToEmoji(),
			Default:     option.Default,
		}
	}

	return &discord.StringSelectComponent{
		CustomID:    c.customID(opts),
		Options:     options,
		Placeholder: c.Placeholder,
		ValueLimits: [2]int{c.MinValues, c.MaxValues},
		Disabled:    c.Disabled,
	}
}

func (c *ComponentData) customID(opts ConvertOptions) discord.ComponentID {
	if opts.ComponentIDFactory != nil {
		return opts.ComponentIDFactory(c)
	}
	return discord.ComponentID(c.FlowSourceID)
}

func (m *UnfurledMediaItemData) toUnfurledMediaItem() discord.UnfurledMediaitem {
	if m == nil {
		return discord.UnfurledMediaitem{}
	}

	return discord.UnfurledMediaitem{
		URL: m.URL,
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
