package store

import (
	"context"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type AppUsageStore interface {
	CreateAppUsageEntry(ctx context.Context, entry model.AppUsageEntry) error
	GetLastAppUsageEntry(ctx context.Context, appID distype.Snowflake) (*model.AppUsageEntry, error)
	GetAppUsageSummary(ctx context.Context, appID distype.Snowflake) (*model.AppUsageSummary, error)
	GetAppUsageAndLimits(ctx context.Context, appID distype.Snowflake) (*model.AppUsageAndLimits, error)
}
