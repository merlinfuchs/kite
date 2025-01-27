package billing

import (
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func (h *BillingHandler) HandleEntitlementList(c *handler.Context) (*wire.EntitlementListResponse, error) {
	entitlements, err := h.entitlementStore.EntitlementsWithSubscription(c.Context(), c.App.ID)
	if err != nil {
		return nil, err
	}

	res := make(wire.EntitlementListResponse, len(entitlements))
	/* for i, entitlement := range entitlements {
		var featureSet FeatureSet
		for _, feature := range h.config.FeatureSets {
			if feature.ID == entitlement.FeatureSetID {
				featureSet = feature
				break
			}
		}

		res[i] = &wire.Entitlement{
			ID:           entitlement.Entitlement.ID,
			Default:      entitlement.Entitlement.Default,
			Subscription: wire.SubscriptionToWire(entitlement.Subscription),
			FeatureSet:   wire.EntitlementFeatureSetToWire(entitlement.Entitlement.FeatureSet),
			CreatedAt:    entitlement.Entitlement.CreatedAt,
			UpdatedAt:    entitlement.Entitlement.UpdatedAt,
		}
	} */

	return &res, nil
}
