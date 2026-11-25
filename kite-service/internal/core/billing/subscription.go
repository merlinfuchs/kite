package billing

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

// SubscriptionManager keeps our record of a subscription in step with
// LemonSqueezy or Paddle, which replaces it. It lives outside the API handler
// because the reconcile command drives the same logic from the CLI.
type SubscriptionManager struct {
	subscriptionStore store.SubscriptionStore
	entitlementStore  store.EntitlementStore
	planManager       *plan.PlanManager

	// client is nil for reconciling without a Lemon Squeezy API key.
	client *lemonsqueezy.Client
	// paddle is nil without a Paddle API key.
	paddle *paddle.SDK
}

// NewLemonSqueezyClient builds the client the subscription manager and the API
// handler share.
func NewLemonSqueezyClient(apiKey string, signingSecret string) *lemonsqueezy.Client {
	return lemonsqueezy.New(
		lemonsqueezy.WithAPIKey(apiKey),
		lemonsqueezy.WithSigningSecret(signingSecret),
	)
}

// NewReconcileLemonSqueezyClient is NewLemonSqueezyClient for reconciling, nil
// without an API key, so Lemon Squeezy subscriptions are skipped instead of
// failing once its key is gone.
func NewReconcileLemonSqueezyClient(apiKey string, signingSecret string) *lemonsqueezy.Client {
	if apiKey == "" {
		return nil
	}
	return NewLemonSqueezyClient(apiKey, signingSecret)
}

func NewSubscriptionManager(
	subscriptionStore store.SubscriptionStore,
	entitlementStore store.EntitlementStore,
	planManager *plan.PlanManager,
	client *lemonsqueezy.Client,
	paddle *paddle.SDK,
) *SubscriptionManager {
	return &SubscriptionManager{
		subscriptionStore: subscriptionStore,
		entitlementStore:  entitlementStore,
		planManager:       planManager,
		client:            client,
		paddle:            paddle,
	}
}

// Sync stores the subscription and brings its entitlements in line with the
// plan and dates its provider reports.
//
// appID is empty when we cannot tell which app the subscription belongs to,
// which is the case for webhooks that arrive without our checkout metadata and
// for reconciliation. Entitlements are then updated in place rather than
// created, because there is no app to attribute a new one to.
func (m *SubscriptionManager) Sync(ctx context.Context, sub model.Subscription, appID string) (*model.Subscription, error) {
	sub.ID = util.UniqueID()

	// Checked before storing anything, so incomplete data never replaces a
	// stored product
	plan, err := m.plan(sub)
	if err != nil {
		return nil, err
	}

	subscription, err := m.upsert(ctx, sub)
	if errors.Is(err, store.ErrOutdated) {
		// A newer state was stored first, the entitlements follow that one
		subscription, err = m.stored(ctx, sub)
		if err != nil {
			return nil, fmt.Errorf("failed to get newer subscription: %w", err)
		}
		if plan, err = m.plan(*subscription); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to upsert subscription: %w", err)
	}

	entitlement := model.Entitlement{
		ID:             util.UniqueID(),
		Type:           "subscription",
		SubscriptionID: null.StringFrom(subscription.ID),
		AppID:          appID,
		PlanID:         plan.ID,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		EndsAt:         entitlementEndsAt(subscription.RenewsAt, subscription.EndsAt),
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

func (m *SubscriptionManager) upsert(ctx context.Context, sub model.Subscription) (*model.Subscription, error) {
	switch sub.Source {
	case model.SubscriptionSourceLemonSqueezy:
		return m.subscriptionStore.UpsertLemonSqueezySubscription(ctx, sub)
	case model.SubscriptionSourcePaddle:
		return m.subscriptionStore.UpsertPaddleSubscription(ctx, sub)
	}
	return nil, fmt.Errorf("unknown subscription source %q", sub.Source)
}

// stored returns the stored subscription with the provider ID of sub.
func (m *SubscriptionManager) stored(ctx context.Context, sub model.Subscription) (*model.Subscription, error) {
	switch sub.Source {
	case model.SubscriptionSourceLemonSqueezy:
		return m.subscriptionStore.SubscriptionByLemonSqueezyID(ctx, sub.LemonsqueezySubscriptionID.String)
	case model.SubscriptionSourcePaddle:
		return m.subscriptionStore.SubscriptionByPaddleID(ctx, sub.PaddleSubscriptionID.String)
	}
	return nil, fmt.Errorf("unknown subscription source %q", sub.Source)
}

// ErrUnknownPlan is returned for subscriptions to a product that isn't a plan.
var ErrUnknownPlan = errors.New("product is not a plan")

func (m *SubscriptionManager) plan(sub model.Subscription) (*model.Plan, error) {
	plan := m.planManager.PlanBySubscription(&sub)
	if plan == nil {
		productID := cmp.Or(sub.LemonsqueezyProductID.String, sub.PaddleProductID.String)
		return nil, fmt.Errorf("%w: %s product ID %s", ErrUnknownPlan, sub.Source, productID)
	}
	return plan, nil
}

// ChangePlan moves the subscription onto another Lemon Squeezy variant or
// Paddle price, prorated and invoiced right away so the new plan's features
// apply from this moment rather than from the next renewal.
func (m *SubscriptionManager) ChangePlan(
	ctx context.Context,
	subscription *model.Subscription,
	priceID string,
	appID string,
) (*model.Subscription, error) {
	switch subscription.Source {
	case model.SubscriptionSourceLemonSqueezy:
		return m.changeLemonSqueezyPlan(ctx, subscription, priceID, appID)
	case model.SubscriptionSourcePaddle:
		return m.changePaddlePlan(ctx, subscription, priceID, appID)
	}
	return nil, fmt.Errorf("unknown subscription source %q", subscription.Source)
}

func (m *SubscriptionManager) changeLemonSqueezyPlan(
	ctx context.Context,
	subscription *model.Subscription,
	priceID string,
	appID string,
) (*model.Subscription, error) {
	variantID, err := strconv.Atoi(priceID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert variant ID to int: %w", err)
	}

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
	return m.Sync(ctx, SubscriptionFromLemonSqueezy(res.Data.ID, res.Data.Attributes, subscription.UserID), appID)
}

// Refresh syncs the subscription as LemonSqueezy currently reports it.
func (m *SubscriptionManager) Refresh(ctx context.Context, lemonSqueezyID string, userID string, appID string) (*model.Subscription, error) {
	res, _, err := m.client.Subscriptions.Get(ctx, lemonSqueezyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription from LemonSqueezy: %w", err)
	}

	return m.Sync(ctx, SubscriptionFromLemonSqueezy(res.Data.ID, res.Data.Attributes, userID), appID)
}

// SubscriptionFromLemonSqueezy adapts a subscription as returned by the
// LemonSqueezy API, which doesn't report our user ID.
func SubscriptionFromLemonSqueezy(lemonSqueezyID string, sub lemonsqueezy.Subscription, userID string) model.Subscription {
	return model.Subscription{
		UserID:          userID,
		Source:          model.SubscriptionSourceLemonSqueezy,
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
// dates, so an entitlement never outlives the period that has been paid for.
func entitlementEndsAt(renewsAt null.Time, endsAt null.Time) null.Time {
	result := model.PaidUntil(renewsAt, endsAt)
	if !result.Valid {
		// Neither date is set, which means the subscription will not renew and
		// has no remaining paid period - a paused one, for example. Ending the
		// entitlement now keeps it out of ActiveEntitlements.
		result = null.TimeFrom(time.Now().UTC())
	}

	return result
}
