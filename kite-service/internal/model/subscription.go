package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Subscription struct {
	ID                         string
	DisplayName                string
	Source                     SubscriptionSource
	Status                     string
	StatusFormatted            string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	RenewsAt                   time.Time
	TrialEndsAt                null.Time
	EndsAt                     null.Time
	UserID                     string
	LemonsqueezySubscriptionID null.String
	LemonsqueezyCustomerID     null.String
	LemonsqueezyOrderID        null.String
	LemonsqueezyProductID      null.String
	LemonsqueezyVariantID      null.String
}

type SubscriptionSource string

const (
	SubscriptionSourceLemonSqueezy SubscriptionSource = "lemonsqueezy"
)

func (s SubscriptionSource) String() string {
	return string(s)
}
