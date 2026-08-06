package billing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

// SubscriptionManager keeps our record of a subscription in step with
// LemonSqueezy. It lives outside the API handler because the reconcile command
// drives the same logic from the CLI.
type SubscriptionManager struct {
	subscriptionStore store.SubscriptionStore
	entitlementStore  store.EntitlementStore
	planManager       *plan.PlanManager

	client *lemonsqueezy.Client
}

// NewLemonSqueezyClient builds the client the subscription manager and the API
// handler share.
func NewLemonSqueezyClient(apiKey string, signingSecret string) *lemonsqueezy.Client {
	return lemonsqueezy.New(
		lemonsqueezy.WithAPIKey(apiKey),
		lemonsqueezy.WithSigningSecret(signingSecret),
	)
}

func NewSubscriptionManager(
	subscriptionStore store.SubscriptionStore,
	entitlementStore store.EntitlementStore,
	planManager *plan.PlanManager,
	client *lemonsqueezy.Client,
) *SubscriptionManager {
	return &SubscriptionManager{
		subscriptionStore: subscriptionStore,
		entitlementStore:  entitlementStore,
		planManager:       planManager,
		client:            client,
	}
}

// Sync stores the subscription and brings its entitlements in line with the
// plan and dates LemonSqueezy reports.
//
// appID is empty when we cannot tell which app the subscription belongs to,
// which is the case for webhooks that arrive without our checkout metadata and
// for reconciliation. Entitlements are then updated in place rather than
// created, because there is no app to attribute a new one to.
func (m *SubscriptionManager) Sync(ctx context.Context, sub model.Subscription, appID string) (*model.Subscription, error) {
	sub.ID = util.UniqueID()
	sub.Source = model.SubscriptionSourceLemonSqueezy

	subscription, err := m.subscriptionStore.UpsertLemonSqueezySubscription(ctx, sub)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert subscription: %w", err)
	}

	productID := sub.LemonsqueezyProductID.String
	plan := m.planManager.PlanByLemonSqueezyProductID(productID)
	if plan == nil {
		return nil, fmt.Errorf("failed to find plan for product ID %s", productID)
	}

	entitlement := model.Entitlement{
		ID:             util.UniqueID(),
		Type:           "subscription",
		SubscriptionID: null.StringFrom(subscription.ID),
		AppID:          appID,
		PlanID:         plan.ID,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		EndsAt:         entitlementEndsAt(sub.RenewsAt, sub.EndsAt),
	}

	if appID != "" {
		// Create a new entitlement or update the existing one
		if _, err := m.entitlementStore.UpsertSubscriptionEntitlement(ctx, entitlement); err != nil {
			return nil, fmt.Errorf("failed to upsert subscription entitlement: %w", err)
		}
	} else {
		// We don't have the app ID, but there might be an entitlement anyway, so we update that
		if err := m.entitlementStore.UpdateSubscriptionEntitlement(ctx, entitlement); err != nil {
			return nil, fmt.Errorf("failed to update subscription entitlement: %w", err)
		}
	}

	return subscription, nil
}

// ChangePlan moves the subscription onto another variant, prorated and invoiced
// right away so the new plan's features apply from this moment rather than from
// the next renewal.
func (m *SubscriptionManager) ChangePlan(
	ctx context.Context,
	subscription *model.Subscription,
	variantID int,
	appID string,
) (*model.Subscription, error) {
	res, _, err := m.client.Subscriptions.Update(ctx, &lemonsqueezy.SubscriptionUpdateParams{
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
	updated := SubscriptionFromLemonSqueezy(res.Data.ID, res.Data.Attributes)
	updated.UserID = subscription.UserID

	return m.Sync(ctx, updated, appID)
}

// SubscriptionFromLemonSqueezy adapts a subscription as returned by the
// LemonSqueezy API. The caller sets UserID, which the API does not report.
func SubscriptionFromLemonSqueezy(lemonSqueezyID string, sub lemonsqueezy.Subscription) model.Subscription {
	return model.Subscription{
		DisplayName:     sub.ProductName,
		Status:          sub.Status,
		StatusFormatted: sub.StatusFormatted,
		// The client decodes renews_at into a plain time.Time, so a null from
		// LemonSqueezy arrives as the zero value rather than as an absent date.
		RenewsAt:                   null.NewTime(sub.RenewsAt, !sub.RenewsAt.IsZero()),
		TrialEndsAt:                null.TimeFromPtr(sub.TrialEndsAt),
		EndsAt:                     null.TimeFromPtr(sub.EndsAt),
		CreatedAt:                  sub.CreatedAt,
		UpdatedAt:                  sub.UpdatedAt,
		LemonsqueezySubscriptionID: null.StringFrom(lemonSqueezyID),
		LemonsqueezyCustomerID:     null.StringFrom(strconv.Itoa(sub.CustomerID)),
		LemonsqueezyOrderID:        null.StringFrom(strconv.Itoa(sub.OrderID)),
		LemonsqueezyProductID:      null.StringFrom(strconv.Itoa(sub.ProductID)),
		LemonsqueezyVariantID:      null.StringFrom(strconv.Itoa(sub.VariantID)),
	}
}

// entitlementEndsAt derives the entitlement end date from the subscription's
// dates. renews_at is the default, so an entitlement never outlives the period
// that has been paid for. ends_at wins when it comes first, which is the case
// for a cancelled subscription serving out its remaining period.
func entitlementEndsAt(renewsAt null.Time, endsAt null.Time) null.Time {
	result := renewsAt
	if endsAt.Valid && (!result.Valid || endsAt.Time.Before(result.Time)) {
		result = endsAt
	}
	if !result.Valid {
		// Neither date is set, which means the subscription will not renew and
		// has no remaining paid period - a paused one, for example. Ending the
		// entitlement now keeps it out of ActiveEntitlements.
		result = null.TimeFrom(time.Now().UTC())
	}

	return result
}
