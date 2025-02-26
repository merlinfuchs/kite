package plan

import (
	"context"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type PlanManagerConfig struct {
	DiscordBotToken string
	DiscordGuildID  string
}

type PlanManager struct {
	entitlementStore  store.EntitlementStore
	subscriptionStore store.SubscriptionStore
	userStore         store.UserStore
	plans             []model.Plan

	config PlanManagerConfig
}

func NewPlanManager(
	entitlementStore store.EntitlementStore,
	subscriptionStore store.SubscriptionStore,
	userStore store.UserStore,
	plans []model.Plan,
	config PlanManagerConfig,
) *PlanManager {
	return &PlanManager{
		entitlementStore:  entitlementStore,
		subscriptionStore: subscriptionStore,
		userStore:         userStore,
		plans:             plans,
		config:            config,
	}
}

func (m *PlanManager) Plans() []model.Plan {
	return m.plans
}

func (m *PlanManager) PlanByLemonSqueezyProductID(productID string) *model.Plan {
	for _, plan := range m.plans {
		if plan.LemonSqueezyProductID == productID {
			return &plan
		}
	}
	return nil
}

func (m *PlanManager) AppFeatures(ctx context.Context, appID string) model.Features {
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
			continue
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
