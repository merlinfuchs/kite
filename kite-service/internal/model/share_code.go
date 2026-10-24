package model

import (
	"encoding/json"
	"time"

	"gopkg.in/guregu/null.v4"
)

type ShareCode struct {
	Code          string
	Type          ShareCodeType
	Data          json.RawMessage
	CreatorUserID string
	AppID         null.String
	CreatedAt     time.Time
	LastUsedAt    time.Time
}

// ShareCodeType says what Data holds, the server never looks inside it.
type ShareCodeType string

const (
	ShareCodeTypeCommand       ShareCodeType = "command"
	ShareCodeTypeEventListener ShareCodeType = "event_listener"
	ShareCodeTypeMessage       ShareCodeType = "message"
	// ShareCodeTypeBundle holds several items in the format of a bundle. Like
	// the others it's a copy, importing it doesn't link to the app it came
	// from.
	ShareCodeTypeBundle ShareCodeType = "bundle"
)
