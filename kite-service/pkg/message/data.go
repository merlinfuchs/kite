package message

import (
	"encoding/json"
	"time"
)

type MessageData struct {
	Content         string               `json:"content,omitempty"`
	Flags           int                  `json:"flags,omitempty"`
	Attachments     []MessageAttachment  `json:"attachments,omitempty"`
	Embeds          []EmbedData          `json:"embeds,omitempty"`
	Components      []ComponentData      `json:"components,omitempty"`
	AllowedMentions *AllowedMentionsData `json:"allowed_mentions,omitempty"`
}

func (m *MessageData) UnmarshalJSON(b []byte) error {
	type rawMessageData MessageData
	if err := json.Unmarshal(b, (*rawMessageData)(m)); err != nil {
		return err
	}

	// Messages saved before components v2 stored action rows without a type.
	for i := range m.Components {
		if m.Components[i].Type == 0 {
			m.Components[i].Type = ComponentTypeActionRow
		}
	}

	return nil
}

func (m *MessageData) EachString(replace func(s *string) error) error {
	if err := replace(&m.Content); err != nil {
		return err
	}

	for e := range m.Embeds {
		embed := &m.Embeds[e]

		if err := replace(&embed.Description); err != nil {
			return err
		}

		if err := replace(&embed.Title); err != nil {
			return err
		}

		if err := replace(&embed.URL); err != nil {
			return err
		}

		if embed.Author != nil {
			if err := replace(&embed.Author.Name); err != nil {
				return err
			}

			if err := replace(&embed.Author.URL); err != nil {
				return err
			}

			if err := replace(&embed.Author.IconURL); err != nil {
				return err
			}

			if embed.Author.Name == "" {
				embed.Author = nil
			}
		}

		if embed.Footer != nil {
			if err := replace(&embed.Footer.Text); err != nil {
				return err
			}

			if err := replace(&embed.Footer.IconURL); err != nil {
				return err
			}

			if embed.Footer.Text == "" {
				embed.Footer = nil
			}
		}

		if embed.Image != nil {
			if err := replace(&embed.Image.URL); err != nil {
				return err
			}

			if embed.Image.URL == "" {
				embed.Image = nil
			}
		}

		if embed.Thumbnail != nil {
			if err := replace(&embed.Thumbnail.URL); err != nil {
				return err
			}

			if embed.Thumbnail.URL == "" {
				embed.Thumbnail = nil
			}
		}

		for f := range embed.Fields {
			field := &embed.Fields[f]

			if err := replace(&field.Name); err != nil {
				return err
			}

			if err := replace(&field.Value); err != nil {
				return err
			}
		}
	}

	return m.EachComponent(func(c *ComponentData) error {
		for _, s := range []*string{&c.Label, &c.Placeholder, &c.URL, &c.Content, &c.Description} {
			if err := replace(s); err != nil {
				return err
			}
		}

		for _, media := range []*UnfurledMediaItemData{c.Media, c.File} {
			if media != nil {
				if err := replace(&media.URL); err != nil {
					return err
				}
			}
		}

		for i := range c.Items {
			item := &c.Items[i]
			if err := replace(&item.Media.URL); err != nil {
				return err
			}
			if err := replace(&item.Description); err != nil {
				return err
			}
		}

		for i := range c.Options {
			option := &c.Options[i]
			if err := replace(&option.Label); err != nil {
				return err
			}
			if err := replace(&option.Description); err != nil {
				return err
			}
		}

		return nil
	})
}

// EachComponent calls fn for every component in the message, including nested children and section accessories.
func (m *MessageData) EachComponent(fn func(c *ComponentData) error) error {
	for i := range m.Components {
		if err := m.Components[i].walk(fn); err != nil {
			return err
		}
	}
	return nil
}

func (c *ComponentData) walk(fn func(c *ComponentData) error) error {
	if err := fn(c); err != nil {
		return err
	}

	for i := range c.Components {
		if err := c.Components[i].walk(fn); err != nil {
			return err
		}
	}

	if c.Accessory != nil {
		if err := c.Accessory.walk(fn); err != nil {
			return err
		}
	}

	return nil
}

// IsComponentsV2 reports whether the message uses the components v2 layout, either through the flag or because it contains a v2-only top-level component.
func (m *MessageData) IsComponentsV2() bool {
	if m.Flags&FlagIsComponentsV2 != 0 {
		return true
	}

	for _, c := range m.Components {
		if c.Type != ComponentTypeActionRow {
			return true
		}
	}

	return false
}

// HasInteractiveComponents reports whether the message contains any component that can trigger an interaction.
func (m *MessageData) HasInteractiveComponents() bool {
	found := false
	m.EachComponent(func(c *ComponentData) error {
		if c.IsInteractive() {
			found = true
		}
		return nil
	})
	return found
}

type MessageAttachment struct {
	AssetID string `json:"asset_id,omitempty"`
}

type EmbedData struct {
	ID int `json:"id,omitempty"`

	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Timestamp   *time.Time          `json:"timestamp,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Footer      *EmbedFooterData    `json:"footer,omitempty"`
	Image       *EmbedImageData     `json:"image,omitempty"`
	Thumbnail   *EmbedThumbnailData `json:"thumbnail,omitempty"`
	Author      *EmbedAuthorData    `json:"author,omitempty"`
	Fields      []EmbedFieldData    `json:"fields,omitempty"`
}

type EmbedFooterData struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type EmbedImageData struct {
	URL string `json:"url,omitempty"`
}

type EmbedThumbnailData struct {
	URL string `json:"url,omitempty"`
}

type EmbedAuthorData struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type EmbedFieldData struct {
	ID int `json:"id,omitempty"`

	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}

const FlagIsComponentsV2 = 1 << 15

const (
	ComponentTypeActionRow    = 1
	ComponentTypeButton       = 2
	ComponentTypeStringSelect = 3
	ComponentTypeSection      = 9
	ComponentTypeTextDisplay  = 10
	ComponentTypeThumbnail    = 11
	ComponentTypeMediaGallery = 12
	ComponentTypeFile         = 13
	ComponentTypeSeparator    = 14
	ComponentTypeContainer    = 17
)

const ButtonStyleLink = 5

type ComponentData struct {
	ID int `json:"id,omitempty"`

	Type     int  `json:"type,omitempty"`
	Disabled bool `json:"disabled,omitempty"`

	// Button
	Style int                 `json:"style,omitempty"`
	Label string              `json:"label,omitempty"`
	Emoji *ComponentEmojiData `json:"emoji,omitempty"`
	URL   string              `json:"url,omitempty"`

	// Select Menu
	Placeholder string                      `json:"placeholder,omitempty"`
	MinValues   int                         `json:"min_values,omitempty"`
	MaxValues   int                         `json:"max_values,omitempty"`
	Options     []ComponentSelectOptionData `json:"options,omitempty"`

	// Action Row, Section, Container
	Components []ComponentData `json:"components,omitempty"`

	// Section
	Accessory *ComponentData `json:"accessory,omitempty"`

	// Text Display
	Content string `json:"content,omitempty"`

	// Thumbnail
	Media       *UnfurledMediaItemData `json:"media,omitempty"`
	Description string                 `json:"description,omitempty"`

	// Thumbnail, File, Container
	Spoiler bool `json:"spoiler,omitempty"`

	// Media Gallery
	Items []MediaGalleryItemData `json:"items,omitempty"`

	// File
	File *UnfurledMediaItemData `json:"file,omitempty"`

	// Separator
	Divider *bool `json:"divider,omitempty"`
	Spacing int   `json:"spacing,omitempty"`

	// Container
	AccentColor *int `json:"accent_color,omitempty"`

	FlowSourceID string `json:"flow_source_id,omitempty"`
}

func (c *ComponentData) IsInteractive() bool {
	switch c.Type {
	case ComponentTypeButton:
		return c.Style != ButtonStyleLink
	case ComponentTypeStringSelect:
		return true
	}
	return false
}

type UnfurledMediaItemData struct {
	URL string `json:"url"`
}

type MediaGalleryItemData struct {
	Media       UnfurledMediaItemData `json:"media"`
	Description string                `json:"description,omitempty"`
	Spoiler     bool                  `json:"spoiler,omitempty"`
}

type ComponentSelectOptionData struct {
	ID int `json:"id,omitempty"`

	Label       string              `json:"label,omitempty"`
	Description string              `json:"description,omitempty"`
	Emoji       *ComponentEmojiData `json:"emoji,omitempty"`
	Default     bool                `json:"default,omitempty"`

	FlowSourceID string `json:"flow_source_id,omitempty"`
}

type ComponentEmojiData struct {
	Name     string `json:"name,omitempty"`
	ID       string `json:"id,omitempty"`
	Animated bool   `json:"animated,omitempty"`
}

type AllowedMentionsData struct {
	Parse []string `json:"parse,omitempty"`
}
