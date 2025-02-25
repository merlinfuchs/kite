package feature

import (
	"context"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type Manager struct {
	entitlementStore store.EntitlementStore
	plans            []model.Plan
}

func NewManager(entitlementStore store.EntitlementStore, plans []model.Plan) *Manager {
	return &Manager{
		entitlementStore: entitlementStore,
		plans:            plans,
	}
}

func (m *Manager) Plans() []model.Plan {
	return m.plans
}

func (m *Manager) PlanByLemonSqueezyProductID(productID string) *model.Plan {
	for _, plan := range m.plans {
		if plan.LemonSqueezyProductID == productID {
			return &plan
		}
	}
	return nil
}

func (m *Manager) AppFeatures(ctx context.Context, appID string) model.Features {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	entitlements, err := m.entitlementStore.ActiveEntitlements(ctx, appID, time.Now().UTC())
	if err != nil {
		slog.Error(
			"Failed to get active entitlements",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
	}

	var features model.Features
	for _, plan := range m.plans {
		if plan.Default {
			features = features.Merge(plan.Features())
		}

		for _, entitlement := range entitlements {
			if entitlement.PlanID == plan.ID {
				features = features.Merge(plan.Features())
				break
			}
		}
	}

	return features
}
