package model

import (
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type EventListenerType string

const (
	EventListenerTypeDiscordMessageCreate EventListenerType = "discord_message_create"
)

func EventTypeFromDiscordEventType(eventType ws.EventType) EventListenerType {
	return EventListenerType("discord_" + strings.ToLower(string(eventType)))
}

type EventListener struct {
	ID            string
	Name          string
	Description   string
	Enabled       bool
	AppID         string
	ModuleID      null.String
	CreatorUserID string
	Type          EventListenerType
	Filter        *EventListenerFilter
	FlowSource    flow.FlowData
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type EventListenerFilter struct{}
