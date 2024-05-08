package model

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"gopkg.in/guregu/null.v4"
)

type AppEntitlementSource string

const (
	AppEntitlementSourceManual  AppEntitlementSource = "MANUAL"
	AppEntitlementSourceDiscord AppEntitlementSource = "DISCORD"
	AppEntitlementSourceDefault AppEntitlementSource = "DEFAULT"
)

type AppEntitlement struct {
	ID          string
	AppID       distype.Snowflake
	UserID      distype.Nullable[distype.Snowflake]
	Source      AppEntitlementSource
	SourceID    null.String
	Name        null.String
	Description null.String
	Features    AppEntitlementFeatures
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ValidFrom   null.Time
	ValidUntil  null.Time
}

type AppEntitlementFeatures struct {
	MonthlyExecutionTimeLimit    time.Duration
	MonthlyExecutionTimeAdditive bool
}

type AppEntitlementResolved struct {
	MonthlyExecutionTimeLimit time.Duration
}
