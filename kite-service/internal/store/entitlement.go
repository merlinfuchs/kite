package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type EntitlementStore interface {
	Entitlements(ctx context.Context, appID string) ([]*model.Entitlement, error)
	UpsertSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) (*model.Entitlement, error)
}
