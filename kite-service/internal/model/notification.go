package model

import (
	"encoding/json"
	"time"

	"gopkg.in/guregu/null.v4"
)

// Notification is something Kite told the owner of an app or a user, for the
// dashboard. It's about the app if AppID is set, otherwise about the user.
type Notification struct {
	ID     string
	AppID  null.String
	UserID null.String
	Type   NotificationType
	// DedupeKey makes the notification be sent only once with the same type,
	// like once a month. Without one it's always sent.
	DedupeKey null.String
	Data      json.RawMessage
	CreatedAt time.Time
}

type NotificationType string

const (
	// NotificationTypeCreditsWarning is sent once a month when an app used 80%
	// of its credits.
	NotificationTypeCreditsWarning NotificationType = "credits_warning"
	// NotificationTypeCreditsPaused is sent when an app is paused for using
	// all of its credits.
	NotificationTypeCreditsPaused NotificationType = "credits_paused"
)

// CreditsNotificationData is the data of the credit notifications.
type CreditsNotificationData struct {
	CreditsUsed  int       `json:"credits_used"`
	CreditsLimit int       `json:"credits_limit"`
	ResetAt      time.Time `json:"reset_at"`
}
