package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Subscription struct {
	ID              string
	DisplayName     string
	Source          SubscriptionSource
	Status          string
	StatusFormatted string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	// RenewsAt is unset for subscriptions that will not renew, such as paused
	// or expired ones.
	RenewsAt                   null.Time
	TrialEndsAt                null.Time
	EndsAt                     null.Time
	UserID                     string
	LemonsqueezySubscriptionID null.String
	LemonsqueezyCustomerID     null.String
	LemonsqueezyOrderID        null.String
	LemonsqueezyProductID      null.String
	LemonsqueezyVariantID      null.String
}

// The statuses a LemonSqueezy subscription can have.
const (
	SubscriptionStatusOnTrial   = "on_trial"
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPastDue   = "past_due"
	SubscriptionStatusUnpaid    = "unpaid"
	SubscriptionStatusCancelled = "cancelled"
	SubscriptionStatusExpired   = "expired"
	SubscriptionStatusPaused    = "paused"
)

// IsActive reports whether the subscription still grants access, or will once a
// failed payment is recovered. The remaining statuses leave the app without an
// entitlement, so the plan has to be purchasable again and cannot be switched.
func (s *Subscription) IsActive() bool {
	switch s.Status {
	case SubscriptionStatusOnTrial, SubscriptionStatusActive, SubscriptionStatusPastDue:
		return true
	default:
		return false
	}
}

type SubscriptionSource string

const (
	SubscriptionSourceLemonSqueezy SubscriptionSource = "lemonsqueezy"
)

func (s SubscriptionSource) String() string {
	return string(s)
}
