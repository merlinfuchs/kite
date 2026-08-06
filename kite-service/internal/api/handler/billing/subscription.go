package billing

import (
	"fmt"
	"strconv"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
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

// activeStatuses are the LemonSqueezy statuses a subscription can be switched
// from. Cancelled, expired, unpaid and paused subscriptions are not billed on a
// schedule anymore, so a plan change would not take effect; those go back
// through checkout or the customer portal instead.
var activeStatuses = map[string]bool{
	"on_trial": true,
	"active":   true,
	"past_due": true,
}

func (h *BillingHandler) HandleSubscriptionPlanUpdate(c *handler.Context, req wire.SubscriptionPlanUpdateRequest) (*wire.SubscriptionPlanUpdateResponse, error) {
	// Scoped to the app so that a subscription entitling some other app cannot
	// be made to grant an entitlement here as well.
	subscriptions, err := h.subscriptionStore.SubscriptionsByAppID(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	subscriptionID := c.Param("subscriptionID")
	var subscription *model.Subscription
	for _, s := range subscriptions {
		if s.ID == subscriptionID {
			subscription = s
			break
		}
	}
	if subscription == nil {
		return nil, handler.ErrNotFound("unknown_subscription", "The app has no such subscription")
	}

	if subscription.UserID != c.Session.UserID {
		return nil, handler.ErrForbidden("missing_access", "You do not have access to this subscription")
	}

	if !subscription.LemonsqueezySubscriptionID.Valid {
		return nil, handler.ErrNotFound("unmanageable_subscription", "Subscription can not be managed")
	}

	if !activeStatuses[subscription.Status] {
		return nil, handler.ErrBadRequest(
			"inactive_subscription",
			"Only an active subscription can be switched to a different plan",
		)
	}

	plan := h.planManager.PlanByLemonSqueezyVariantID(req.LemonSqueezyVariantID)
	if plan == nil || plan.Hidden {
		return nil, handler.ErrBadRequest("unknown_plan", "There is no plan with that variant ID")
	}

	if subscription.LemonsqueezyVariantID.String == req.LemonSqueezyVariantID {
		return nil, handler.ErrBadRequest("plan_unchanged", "The subscription is already on that plan")
	}

	variantID, err := strconv.Atoi(req.LemonSqueezyVariantID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert variant ID to int: %w", err)
	}

	// Prorated and invoiced right away, so the new plan's features apply from
	// this moment rather than from the next renewal.
	res, _, err := h.client.Subscriptions.Update(c.Context(), &lemonsqueezy.SubscriptionUpdateParams{
		ID: subscription.LemonsqueezySubscriptionID.String,
		Attributes: lemonsqueezy.SubscriptionUpdateParamsAttributes{
			VariantID:          variantID,
			InvoiceImmediately: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription in LemonSqueezy: %w", err)
	}

	// LemonSqueezy also sends a subscription_updated webhook, but we cannot rely
	// on it having arrived by the time the client refetches.
	updated, err := h.syncSubscription(c.Context(), subscriptionSyncFromLemonSqueezy(
		res.Data.ID,
		subscription.UserID,
		c.App.ID,
		res.Data.Attributes,
	))
	if err != nil {
		return nil, err
	}

	return wire.SubscriptionToWire(updated, c.Session.UserID), nil
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
