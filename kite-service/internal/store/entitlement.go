package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type EntitlementStore interface {
	Entitlements(ctx context.Context, appID string) ([]*model.Entitlement, error)
	ActiveEntitlements(ctx context.Context, appID string, now time.Time) ([]*model.Entitlement, error)
	UpsertSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) (*model.Entitlement, error)
	UpdateSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) (*model.Entitlement, error)
}
