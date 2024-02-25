package model

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"gopkg.in/guregu/null.v4"
)

type GuildEntitlementSource string

const (
	GuildEntitlementSourceManual  GuildEntitlementSource = "MANUAL"
	GuildEntitlementSourceDiscord GuildEntitlementSource = "DISCORD"
	GuildEntitlementSourceDefault GuildEntitlementSource = "DEFAULT"
)

type GuildEntitlement struct {
	ID          string
	GuildID     distype.Snowflake
	UserID      distype.Nullable[distype.Snowflake]
	Source      GuildEntitlementSource
	SourceID    null.String
	Name        null.String
	Description null.String
	Features    GuildEntitlementFeatures
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ValidFrom   null.Time
	ValidUntil  null.Time
}

type GuildEntitlementFeatures struct {
	MonthlyCpuTimeLimit    time.Duration
	MonthlyCpuTimeAdditive bool
}

type GuildEntitlementResolved struct {
	MonthlyCpuTimeLimit int
}
