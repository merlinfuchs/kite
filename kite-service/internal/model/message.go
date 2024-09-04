package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"gopkg.in/guregu/null.v4"
)

type Message struct {
	ID            string
	Name          string
	Description   null.String
	AppID         string
	ModuleID      null.String
	CreatorUserID string
	Data          message.MessageData
	FlowSources   map[string]flow.FlowData
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type MessageInstance struct {
	ID               uint64
	MessageID        string
	DiscordGuildID   string
	DiscordChannelID string
	DiscordMessageID string
	FlowSources      map[string]flow.FlowData
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
