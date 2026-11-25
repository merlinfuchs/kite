package billing

import (
	"fmt"
	"log/slog"

	"github.com/PaddleHQ/paddle-go-sdk/v6"
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
		res[i] = wire.SubscriptionToWire(subscription, c.Session.UserID, h.planManager.PlanBySubscription(subscription))
	}

	return &res, nil
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

	if subscription.ProviderID() == "" {
		return nil, handler.ErrNotFound("unmanageable_subscription", "Subscription can not be managed")
	}

	// Cancelled, expired, unpaid and paused subscriptions are not billed on a
	// schedule anymore, so a plan change would not take effect; those go back
	// through checkout or the customer portal instead.
	if !subscription.IsActive() {
		return nil, handler.ErrBadRequest(
			"inactive_subscription",
			"Only an active subscription can be switched to a different plan",
		)
	}

	// The subscription stays with its provider, so the plan has to be sold
	// there too.
	plan := h.planManager.PlanByID(req.PlanID)
	if plan == nil || plan.Hidden {
		return nil, handler.ErrBadRequest("unknown_plan", "There is no such plan")
	}
	priceID := plan.PriceID(subscription.Source, req.Interval == wire.BillingIntervalYearly)
	if priceID == "" {
		return nil, handler.ErrBadRequest("unknown_plan", fmt.Sprintf("The subscription can't be switched to the plan %s", req.Interval))
	}

	if subscription.PriceID() == priceID {
		return nil, handler.ErrBadRequest("plan_unchanged", "The subscription is already on that plan")
	}

	updated, err := h.subscriptionManager.ChangePlan(c.Context(), subscription, priceID, c.App.ID)
	if err != nil {
		// Unexpected errors are returned to the client verbatim, so keep the
		// provider's name out of a message the user will read.
		slog.Error(
			"Failed to switch the subscription's plan",
			slog.String("subscription_id", subscription.ID),
			slog.String("provider_subscription_id", subscription.ProviderID()),
			slog.String("error", err.Error()),
		)
		return nil, handler.ErrInternal("Failed to switch plan with our payment provider, please try again")
	}

	return wire.SubscriptionToWire(updated, c.Session.UserID, h.planManager.PlanBySubscription(updated)), nil
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

	if subscription.Source == model.SubscriptionSourcePaddle {
		return h.paddleSubscriptionManage(c, subscription)
	}

	if subscription.ProviderID() == "" {
		return nil, handler.ErrNotFound("unmanageable_subscription", "Subscription can not be managed")
	}

	sub, _, err := h.client.Subscriptions.Get(c.Context(), subscription.LemonsqueezySubscriptionID.String)
	if err != nil {
		slog.Error(
			"Failed to get subscription from the billing provider",
			slog.String("subscription_id", subscription.ID),
			slog.String("ls_subscription_id", subscription.LemonsqueezySubscriptionID.String),
			slog.String("error", err.Error()),
		)
		return nil, handler.ErrInternal("Failed to reach our payment provider, please try again")
	}

	return &wire.SubscriptionManageResponse{
		CustomerPortalURL:      sub.Data.Attributes.Urls.CustomerPortal,
		UpdatePaymentMethodURL: sub.Data.Attributes.Urls.UpdatePaymentMethod,
	}, nil
}

// paddleSubscriptionManage opens the subscription in the customer portal of
// Paddle, whose links only work for a while, so they're made on each click.
func (h *BillingHandler) paddleSubscriptionManage(c *handler.Context, subscription *model.Subscription) (*wire.SubscriptionManageResponse, error) {
	if h.paddle == nil || !subscription.PaddleCustomerID.Valid || !subscription.PaddleSubscriptionID.Valid {
		return nil, handler.ErrNotFound("unmanageable_subscription", "Subscription can not be managed")
	}

	session, err := h.paddle.CreateCustomerPortalSession(c.Context(), &paddle.CreateCustomerPortalSessionRequest{
		CustomerID:      subscription.PaddleCustomerID.String,
		SubscriptionIDs: []string{subscription.PaddleSubscriptionID.String},
	})
	if err != nil {
		slog.Error(
			"Failed to create Paddle customer portal session",
			slog.String("subscription_id", subscription.ID),
			slog.String("paddle_subscription_id", subscription.PaddleSubscriptionID.String),
			slog.String("error", err.Error()),
		)
		return nil, handler.ErrInternal("Failed to reach our payment provider, please try again")
	}

	res := &wire.SubscriptionManageResponse{CustomerPortalURL: session.URLs.General.Overview}
	for _, urls := range session.URLs.Subscriptions {
		if urls.ID == subscription.PaddleSubscriptionID.String {
			res.UpdatePaymentMethodURL = urls.UpdateSubscriptionPaymentMethod
		}
	}
	return res, nil
}
