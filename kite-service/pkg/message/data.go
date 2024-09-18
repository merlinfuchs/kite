package message

import (
	"time"
)

type MessageData struct {
	Content     string              `json:"content,omitempty"`
	Flags       int                 `json:"flags,omitempty"`
	Attachments []MessageAttachment `json:"attachments,omitempty"`
	Embeds      []EmbedData         `json:"embeds,omitempty"`
	Components  []ComponentRowData  `json:"components,omitempty"`
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
