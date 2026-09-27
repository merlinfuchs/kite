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
)

type EventListenerType string

const (
	EventListenerTypeDiscordMessageCreate       EventListenerType = "message_create"
	EventListenerTypeDiscordMessageUpdate       EventListenerType = "message_update"
	EventListenerTypeDiscordMessageDelete       EventListenerType = "message_delete"
	EventListenerTypeDiscordGuildMemberAdd      EventListenerType = "guild_member_add"
	EventListenerTypeDiscordGuildMemberRemove   EventListenerType = "guild_member_remove"
	EventListenerTypeDiscordChannelCreate       EventListenerType = "channel_create"
	EventListenerTypeDiscordChannelUpdate       EventListenerType = "channel_update"
	EventListenerTypeDiscordChannelDelete       EventListenerType = "channel_delete"
	EventListenerTypeDiscordChannelPinsUpdate   EventListenerType = "channel_pins_update"
	EventListenerTypeDiscordThreadCreate        EventListenerType = "thread_create"
	EventListenerTypeDiscordThreadUpdate        EventListenerType = "thread_update"
	EventListenerTypeDiscordThreadDelete        EventListenerType = "thread_delete"
	EventListenerTypeDiscordThreadMembersUpdate EventListenerType = "thread_members_update"

	EventListenerTypeScheduleCron EventListenerType = EventListenerType(flow.EventTypeScheduleCron)
)

// EventSourceForType derives the source from the type, since the type lives in
// the flow and can change with every save while the source is stored separately.
func EventSourceForType(t EventListenerType) EventSource {
	if t == EventListenerTypeScheduleCron {
		return EventSourceSchedule
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
}

type EventListenerFilter struct{}
