package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Entitlement struct {
	ID             string
	Type           string
	SubscriptionID null.String
	AppID          string
	PlanID         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EndsAt         null.Time
}

type EntitlementWithSubscription struct {
	Entitlement
	Subscription *Subscription
}
