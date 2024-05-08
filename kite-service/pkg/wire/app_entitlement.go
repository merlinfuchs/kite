package wire

import "github.com/merlinfuchs/kite/kite-service/pkg/model"

// AppEntitlementResolved represents the resolved entitlements for an app
// All times are in milliseconds
type AppEntitlementResolved struct {
	MonthlyExecutionTimeLimit int `json:"monthly_execution_time_limit"`
}

type AppEntitlementResolvedGetResponse APIResponse[AppEntitlementResolved]

func AppEntitlementResolvedToWire(d *model.AppEntitlementResolved) AppEntitlementResolved {
	return AppEntitlementResolved{
		MonthlyExecutionTimeLimit: int(d.MonthlyExecutionTimeLimit.Milliseconds()),
	}
}
