package billing

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func (h *BillingHandler) HandleAppSubscriptionList(c *handler.Context) (*wire.SubscriptionListResponse, error) {
	subscriptions, err := h.subscriptionStore.SubscriptionsByAppID(c.Context(), c.App.ID)
	if err != nil {
		return nil, err
	}

	res := make(wire.SubscriptionListResponse, len(subscriptions))
	for i, subscription := range subscriptions {
		res[i] = wire.SubscriptionToWire(subscription, c.Session.UserID)
	}

	return &res, nil
}

func (h *BillingHandler) HandleSubscriptionManage(c *handler.Context) (*wire.SubscriptionManageResponse, error) {
	subscriptionID := c.Param("subscriptionID")
	subscription, err := h.subscriptionStore.Subscription(c.Context(), subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription.UserID != c.Session.UserID {
		return nil, handler.ErrForbidden("missing_access", "You do not have access to this subscription")
	}

	if !subscription.LemonsqueezySubscriptionID.Valid {
		return nil, handler.ErrNotFound("unmanageable_subscription", "Subscription can not be managed")
	}

	sub, _, err := h.client.Subscriptions.Get(c.Context(), subscription.LemonsqueezySubscriptionID.String)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription from LemonSqueezy: %w", err)
	}

	return &wire.SubscriptionManageResponse{
		CustomerPortalURL:      sub.Data.Attributes.Urls.CustomerPortal,
		UpdatePaymentMethodURL: sub.Data.Attributes.Urls.UpdatePaymentMethod,
	}, nil
}
