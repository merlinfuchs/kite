package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Entitlement struct {
	ID                          string
	Type                        string
	SubscriptionID              null.String
	AppID                       string
	FeatureUsageCreditsPerMonth int32
	FeatureMaxCollaborator      int32
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	EndsAt                      null.Time
}
