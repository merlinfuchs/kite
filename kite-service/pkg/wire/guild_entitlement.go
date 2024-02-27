package wire

import "github.com/merlinfuchs/kite/kite-service/pkg/model"

// GuildEntitlementResolved represents the resolved entitlements for a guild
// All times are in milliseconds
type GuildEntitlementResolved struct {
	MonthlyExecutionTimeLimit int `json:"monthly_execution_time_limit"`
}

type GuildEntitlementResolvedGetResponse APIResponse[GuildEntitlementResolved]

func GuildEntitlementResolvedToWire(d *model.GuildEntitlementResolved) GuildEntitlementResolved {
	return GuildEntitlementResolved{
		MonthlyExecutionTimeLimit: int(d.MonthlyExecutionTimeLimit.Milliseconds()),
	}
}
