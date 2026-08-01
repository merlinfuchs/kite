package plan

import (
	"context"
	"fmt"
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

	return m.featuresFromEntitlements(entitlements)
}

// DefaultFeatures returns the features every app gets without any
// entitlement. Because Features.Merge takes the maximum of each field, no
// app's resolved features can be below this, so callers looking for apps that
// exceed a limit can rule out anything under it without a lookup.
func (m *PlanManager) DefaultFeatures() model.Features {
	return m.featuresFromEntitlements(nil)
}

// AppFeaturesForApps resolves features for many apps in one round trip.
//
// Apps with no active entitlements still get the default plan's features, so
// the result has an entry for every requested app.
func (m *PlanManager) AppFeaturesForApps(ctx context.Context, appIDs []string) (map[string]model.Features, error) {
	entitlements, err := m.entitlementStore.ActiveEntitlementsForApps(ctx, appIDs, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to get active entitlements: %w", err)
	}

	features := make(map[string]model.Features, len(appIDs))
	for _, appID := range appIDs {
		features[appID] = m.featuresFromEntitlements(entitlements[appID])
	}

	return features, nil
}

// featuresFromEntitlements merges the default plan with every plan the given
// entitlements grant.
func (m *PlanManager) featuresFromEntitlements(entitlements []*model.Entitlement) model.Features {
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
