package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type EntitlementStore interface {
	Entitlements(ctx context.Context, appID string) ([]*model.Entitlement, error)
	ActiveEntitlements(ctx context.Context, appID string, now time.Time) ([]*model.Entitlement, error)
	// ActiveEntitlementsForApps is the batch form, keyed by app ID.
	ActiveEntitlementsForApps(ctx context.Context, appIDs []string, now time.Time) (map[string][]*model.Entitlement, error)
	UpsertSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) (*model.Entitlement, error)
	// UpdateSubscriptionEntitlement updates every entitlement the subscription
	// holds. A subscription with no entitlements is not an error.
	UpdateSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) error
}
