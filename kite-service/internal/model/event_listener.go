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
	EventSourceDiscord EventSource = "discord"
)

type EventListenerType string

const (
	EventListenerTypeDiscordMessageCreate     EventListenerType = "message_create"
	EventListenerTypeDiscordMessageUpdate     EventListenerType = "message_update"
	EventListenerTypeDiscordMessageDelete     EventListenerType = "message_delete"
	EventListenerTypeDiscordGuildMemberAdd    EventListenerType = "guild_member_add"
	EventListenerTypeDiscordGuildMemberRemove EventListenerType = "guild_member_remove"
)

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
}

type EventListenerFilter struct{}
