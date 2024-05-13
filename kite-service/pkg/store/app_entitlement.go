package store

import (
	"context"
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type AppEntitlementStore interface {
	// UpsertAppEntitlement creates a new entitlement or updates an existing one if (app_id, source, source_id) match
	UpsertAppEntitlement(ctx context.Context, entilement model.AppEntitlement) (*model.AppEntitlement, error)
	GetAppEntitlements(ctx context.Context, appID distype.Snowflake, validAt time.Time) ([]model.AppEntitlement, error)
	GetResolvedAppEntitlement(ctx context.Context, appID distype.Snowflake) (*model.AppEntitlementResolved, error)
}
