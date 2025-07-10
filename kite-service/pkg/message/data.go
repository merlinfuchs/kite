package message

import (
	"time"
)

type MessageData struct {
	Content         string               `json:"content,omitempty"`
	Flags           int                  `json:"flags,omitempty"`
	Attachments     []MessageAttachment  `json:"attachments,omitempty"`
	Embeds          []EmbedData          `json:"embeds,omitempty"`
	Components      []ComponentRowData   `json:"components,omitempty"`
	AllowedMentions *AllowedMentionsData `json:"allowed_mentions,omitempty"`
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

	for c := range m.Components {
		component := &m.Components[c]

		for _, option := range component.Components {
			if err := replace(&option.Label); err != nil {
				return err
			}

			if err := replace(&option.Placeholder); err != nil {
				return err
			}
		}
	}

	return nil
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

type ComponentRowData struct {
	ID int `json:"id,omitempty"`

	Components []ComponentData `json:"components,omitempty"`
}

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

	FlowSourceID string `json:"flow_source_id,omitempty"`
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
