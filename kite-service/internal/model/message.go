package model

import (
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type Message struct {
	ID            string
	Name          string
	Description   null.String
	AppID         string
	ModuleID      null.String
	CreatorUserID string
	Data          MessageData
	FlowSources   map[string]flow.FlowData
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TODO: move this type into separate "message" pkg?
type MessageData struct {
	Content string          `json:"content"`
	Embeds  []discord.Embed `json:"embeds"`
}
