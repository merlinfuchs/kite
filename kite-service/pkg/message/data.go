package message

import "time"

type MessageData struct {
	Content string      `json:"content,omitempty"`
	Flags   int         `json:"flags,omitempty"`
	Embeds  []EmbedData `json:"embeds,omitempty"`
}

type EmbedData struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Timestamp   time.Time           `json:"timestamp,omitempty"`
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
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}
