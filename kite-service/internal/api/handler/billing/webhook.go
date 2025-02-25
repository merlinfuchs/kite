package billing

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

func (h *BillingHandler) HandleBillingWebhook(c *handler.Context, body json.RawMessage) (*wire.BillingWebhookResponse, error) {
	eventName := c.Header("X-Event-Name")
	signature := c.Header("X-Signature")

	if !h.client.Webhooks.Verify(c.Context(), signature, body) {
		fmt.Println("failed to verify webhook signature")
		return nil, fmt.Errorf("failed to verify webhook signature")
	}

	if eventName != lemonsqueezy.WebhookEventSubscriptionCreated && eventName != lemonsqueezy.WebhookEventSubscriptionUpdated {
		return nil, fmt.Errorf("unsupported event name: %s", eventName)
	}

	var req wire.BillingWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook event: %w", err)
	}

	appID, _ := req.Meta.CustomData["app_id"].(string)
	userID, _ := req.Meta.CustomData["user_id"].(string)
	if userID == "" {
		slog.Error(
			"Subscription created webhook received without user_id in metadata",
			slog.String("ls_subscription_id", req.Data.ID),
		)
		return nil, fmt.Errorf("user_id is required in metadata")
	}

	sub := req.Data.Attributes

	subscription, err := h.subscriptionStore.UpsertLemonSqueezySubscription(c.Context(), model.Subscription{
		ID:                         util.UniqueID(),
		DisplayName:                sub.ProductName,
		Source:                     model.SubscriptionSourceLemonSqueezy,
		Status:                     sub.Status,
		StatusFormatted:            sub.StatusFormatted,
		RenewsAt:                   sub.RenewsAt,
		TrialEndsAt:                sub.TrialEndsAt,
		EndsAt:                     sub.EndsAt,
		CreatedAt:                  sub.CreatedAt,
		UpdatedAt:                  sub.UpdatedAt,
		UserID:                     userID,
		LemonsqueezySubscriptionID: null.StringFrom(req.Data.ID),
		LemonsqueezyCustomerID:     null.StringFrom(fmt.Sprintf("%d", sub.CustomerID)),
		LemonsqueezyOrderID:        null.StringFrom(fmt.Sprintf("%d", sub.OrderID)),
		LemonsqueezyProductID:      null.StringFrom(fmt.Sprintf("%d", sub.ProductID)),
		LemonsqueezyVariantID:      null.StringFrom(fmt.Sprintf("%d", sub.VariantID)),
	})
	if err != nil {
		slog.Error(
			"Failed to upsert subscription",
			slog.String("ls_subscription_id", req.Data.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to upsert subscription: %w", err)
	}

	// We use the renews_at date as the default entitlement end date
	// This ensures that the entitlement is invalidated for paused subscriptions
	// If the subscription ends at a date that is before the renews_at date, we use the subscription ends_at date
	entitlementEndsAt := sub.RenewsAt
	if sub.EndsAt.Valid && sub.EndsAt.Time.Before(entitlementEndsAt) {
		entitlementEndsAt = sub.EndsAt.Time
	}

	plan := h.planManager.PlanByLemonSqueezyProductID(fmt.Sprintf("%d", sub.ProductID))
	if plan == nil {
		slog.Error(
			"Failed to find plan ID for subscription",
			slog.String("ls_subscription_id", req.Data.ID),
			slog.String("ls_product_id", fmt.Sprintf("%d", sub.ProductID)),
		)
		return nil, fmt.Errorf("failed to find plan ID for subscription")
	}

	entitlement := model.Entitlement{
		ID:             util.UniqueID(),
		Type:           "subscription",
		SubscriptionID: null.StringFrom(subscription.ID),
		AppID:          appID,
		PlanID:         plan.ID,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		EndsAt:         null.TimeFrom(entitlementEndsAt),
	}

	if appID != "" {
		// Create a new entitlement or update the existing one
		_, err := h.entitlementStore.UpsertSubscriptionEntitlement(c.Context(), entitlement)
		if err != nil {
			slog.Error(
				"Failed to upsert subscription entitlement",
				slog.String("subscription_id", subscription.ID),
				slog.String("ls_subscription_id", req.Data.ID),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to upsert subscription entitlement: %w", err)
		}
	} else {
		// We don't have the app ID, but there might be an entitlement anyway, so we update that
		_, err := h.entitlementStore.UpdateSubscriptionEntitlement(c.Context(), entitlement)
		if err != nil {
			slog.Error(
				"Failed to update subscription entitlement",
				slog.String("subscription_id", subscription.ID),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to update subscription entitlement: %w", err)
		}
	}

	return &wire.BillingWebhookResponse{}, nil
}
