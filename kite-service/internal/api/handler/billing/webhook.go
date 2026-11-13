package billing

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

// subscriptionEventNames are the events whose data object is a subscription,
// and which therefore carry a status transition we need to persist. The
// subscription_payment_* events are deliberately absent: their data object is a
// subscription invoice, so its ID and status belong to the invoice, not the
// subscription.
var subscriptionEventNames = map[string]bool{
	lemonsqueezy.WebhookEventSubscriptionCreated:   true,
	lemonsqueezy.WebhookEventSubscriptionUpdated:   true,
	lemonsqueezy.WebhookEventSubscriptionCancelled: true,
	lemonsqueezy.WebhookEventSubscriptionResumed:   true,
	lemonsqueezy.WebhookEventSubscriptionExpired:   true,
	lemonsqueezy.WebhookEventSubscriptionPaused:    true,
	lemonsqueezy.WebhookEventSubscriptionUnpaused:  true,
}

func (h *BillingHandler) HandleBillingWebhook(c *handler.Context, body json.RawMessage) (*wire.BillingWebhookResponse, error) {
	eventName := c.Header("X-Event-Name")
	signature := c.Header("X-Signature")

	if h.config.LemonSqueezySigningSecret == "" || !h.client.Webhooks.Verify(c.Context(), signature, body) {
		return nil, handler.ErrUnauthorized("invalid_signature", "failed to verify webhook signature")
	}

	if !subscriptionEventNames[eventName] {
		// Everything else (orders, license keys, and the subscription_payment_*
		// events, which carry an invoice rather than a subscription) is
		// acknowledged and ignored. Returning an error here would make
		// LemonSqueezy retry and eventually mark the endpoint as failing.
		return &wire.BillingWebhookResponse{}, nil
	}

	// Only the ID and our metadata are used, the state is fetched from the API
	var req struct {
		Meta struct {
			CustomData map[string]any `json:"custom_data"`
		} `json:"meta"`
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook event: %w", err)
	}

	appID, _ := req.Meta.CustomData["app_id"].(string)
	userID, _ := req.Meta.CustomData["user_id"].(string)
	if userID == "" {
		// Lifecycle events raised outside of checkout (a cancellation from the
		// customer portal, for example) can arrive without the custom data we
		// set at checkout. As long as we have seen the subscription before we
		// can recover the owner from it rather than dropping the event.
		existing, err := h.subscriptionStore.SubscriptionByLemonSqueezyID(c.Context(), req.Data.ID)
		if err != nil {
			slog.Error(
				"Subscription webhook received without user_id in metadata for an unknown subscription",
				slog.String("event_name", eventName),
				slog.String("ls_subscription_id", req.Data.ID),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("user_id is required in metadata: %w", err)
		}

		userID = existing.UserID
	}

	// The payload itself can be stale or incomplete: retries arrive out of
	// order, and right after checkout LemonSqueezy sends one without a product.
	_, err := h.subscriptionManager.Refresh(c.Context(), req.Data.ID, userID, appID)
	if err != nil {
		slog.Error(
			"Failed to sync subscription from webhook",
			slog.String("event_name", eventName),
			slog.String("ls_subscription_id", req.Data.ID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return &wire.BillingWebhookResponse{}, nil
}
