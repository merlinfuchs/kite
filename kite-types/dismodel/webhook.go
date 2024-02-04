package dismodel

import "github.com/merlinfuchs/kite/kite-types/null"

type Webhook struct {
	ID            string      `json:"id"`
	Type          WebhookType `json:"type"`
	GuildID       string      `json:"guild_id"`
	ChannelID     string      `json:"channel_id"`
	User          User        `json:"user"`
	Name          string      `json:"name"`
	Avatar        string      `json:"avatar"`
	Token         string      `json:"token"`
	ApplicationID string      `json:"application_id"`
	SourceGuild   *Guild      `json:"source_guild"`
	SourceCHannel *Channel    `json:"source_channel"`
	URL           string      `json:"url"`
}

type WebhookType int

const (
	WebhookTypeIncoming        WebhookType = 1
	WebhookTypeChannelFollower WebhookType = 2
	WebhookTypeApplication     WebhookType = 3
)

type WebhooksUpdateEvent struct {
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
}

type WebhookCreateCall struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}

type WebhookCreateResponse = Webhook

type WebhookListForChannelCall struct {
	ChannelID string `json:"channel_id"`
}

type WebhookListForChannelResponse = []Webhook

type WebhookListForGuildCall struct{}

type WebhookListForGuildResponse = []Webhook

type WebhookGetCall struct {
	ID string `json:"id"`
}

type WebhookGetResponse = Webhook

type WebhookGetWithTokenCall struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type WebhookGetWithTokenResponse = Webhook

type WebhookUpdateCall struct {
	ID        string            `json:"id"`
	Name      null.Null[string] `json:"name,omitempty"`
	Avatar    null.Null[string] `json:"avatar,omitempty"`
	ChannelID null.Null[string] `json:"channel_id,omitempty"`
}

type WebhookUpdateResponse = Webhook

type WebhookUpdateWithTokenCall struct {
	ID        string            `json:"id"`
	Token     string            `json:"token"`
	Name      null.Null[string] `json:"name,omitempty"`
	Avatar    null.Null[string] `json:"avatar,omitempty"`
	ChannelID null.Null[string] `json:"channel_id,omitempty"`
}

type WebhookUpdateWithTokenResponse = Webhook

type WebhookDeleteCall struct {
	ID string `json:"id"`
}

type WebhookDeleteResponse struct{}

type WebhookDeleteWithTokenCall struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type WebhookDeleteWithTokenResponse struct{}

type WebhookExecuteCall struct {
	ID       string `json:"id"`
	Token    string `json:"token"`
	Wait     bool   `json:"wait,omitempty"`
	ThreadID string `json:"thread_id,omitempty"`
	// TODO: make separate type for all calls that create a message
}

type WebhookExecuteResponse = null.Null[Message]

type WebhookGetMessageCall struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	MessageID string `json:"message_id"`
}

type WebhookGetMessageResponse = Message

type WebhookEditMessageCall struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	MessageID string `json:"message_id"`
	// TODO: make separate type for all calls that create a message
}

type WebhookEditMessageResponse = Message

type WebhookDeleteMessageCall struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	MessageID string `json:"message_id"`
}

type WebhookDeleteMessageResponse struct{}
