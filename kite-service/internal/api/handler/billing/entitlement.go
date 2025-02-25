package billing

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func (h *BillingHandler) HandleEntitlementFeaturesGet(c *handler.Context) (*wire.EntitlementFeaturesGetResponse, error) {
	entitlements, err := h.entitlementStore.Entitlements(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entitlements: %w", err)
	}

	// TODO: move into "entitlement manager"

	var features wire.EntitlementFeatures
	for _, entitlement := range entitlements {
		for _, plan := range h.config.Plans {
			if plan.ID == entitlement.PlanID {
				features.MaxCollaborators = max(features.MaxCollaborators, plan.FeatureMaxCollaborators)
				features.UsageCreditsPerMonth = max(features.UsageCreditsPerMonth, plan.FeatureUsageCreditsPerMonth)
				features.MaxGuilds = max(features.MaxGuilds, plan.FeatureMaxGuilds)
				features.PrioritySupport = features.PrioritySupport || plan.FeaturePrioritySupport
				break
			}
		}
	}

	return &features, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
