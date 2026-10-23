package model

import (
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type EventSource string

const (
	EventSourceDiscord  EventSource = "discord"
	EventSourceSchedule EventSource = "schedule"
	EventSourceWebhook  EventSource = "webhook"
)

type EventListenerType string

const (
	EventListenerTypeDiscordMessageCreate         EventListenerType = "message_create"
	EventListenerTypeDiscordMessageUpdate         EventListenerType = "message_update"
	EventListenerTypeDiscordMessageDelete         EventListenerType = "message_delete"
	EventListenerTypeDiscordGuildMemberAdd        EventListenerType = "guild_member_add"
	EventListenerTypeDiscordGuildMemberRemove     EventListenerType = "guild_member_remove"
	EventListenerTypeDiscordMessageReactionAdd    EventListenerType = "message_reaction_add"
	EventListenerTypeDiscordMessageReactionRemove EventListenerType = "message_reaction_remove"
	EventListenerTypeDiscordGuildCreate           EventListenerType = "guild_create"
	EventListenerTypeDiscordGuildDelete           EventListenerType = "guild_delete"
	EventListenerTypeDiscordDirectMessageCreate   EventListenerType = "direct_message_create"
	// Discord has no gateway event for boosts. This type is derived from the
	// MESSAGE_CREATE of the system message Discord posts when a member boosts.
	EventListenerTypeDiscordGuildBoost EventListenerType = "guild_boost"

	EventListenerTypeScheduleCron EventListenerType = EventListenerType(flow.EventTypeScheduleCron)
	EventListenerTypeWebhook      EventListenerType = EventListenerType(flow.EventTypeWebhook)
)

// EventSourceForType derives the source from the type, since the type lives in
// the flow and can change with every save while the source is stored separately.
func EventSourceForType(t EventListenerType) EventSource {
	switch t {
	case EventListenerTypeScheduleCron:
		return EventSourceSchedule
	case EventListenerTypeWebhook:
		return EventSourceWebhook
	}
	return EventSourceDiscord
}

func EventTypeFromDiscordEventType(eventType ws.EventType) EventListenerType {
	return EventListenerType(strings.ToLower(string(eventType)))
}

type EventListener struct {
	ID            string
	Source        EventSource
	Type          EventListenerType
	Description   string
	Enabled       bool
	AppID         string
	ModuleID      null.String
	CreatorUserID string
	Filter        *EventListenerFilter
	FlowSource    flow.FlowData
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastRunAt     null.Time
	// WebhookSecret is the last part of the webhook URL, only set for webhook
	// listeners. It's kept out of the flow, which is exported and shared.
	WebhookSecret null.String
}

type EventListenerFilter struct{}
