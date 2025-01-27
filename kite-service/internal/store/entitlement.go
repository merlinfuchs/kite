package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type EntitlementStore interface {
	Entitlements(ctx context.Context, appID string) ([]*model.Entitlement, error)
	EntitlementsWithSubscription(ctx context.Context, appID string) ([]*model.EntitlementWithSubscription, error)
	UpsertSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) (*model.Entitlement, error)
	UpdateSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) (*model.Entitlement, error)
}
