package billing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

// subscriptionSync is the subset of a LemonSqueezy subscription we persist. It
// exists because the webhook payload and the API client return the same
// subscription in two different shapes.
type subscriptionSync struct {
	LemonSqueezySubscriptionID string
	UserID                     string
	// AppID is empty when we cannot tell which app the subscription belongs to,
	// which is the case for webhooks that arrive without our checkout metadata.
	AppID string

	ProductName     string
	Status          string
	StatusFormatted string

	LemonSqueezyCustomerID string
	LemonSqueezyOrderID    string
	LemonSqueezyProductID  string
	LemonSqueezyVariantID  string

	RenewsAt    null.Time
	TrialEndsAt null.Time
	EndsAt      null.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// subscriptionSyncFromLemonSqueezy adapts a subscription as returned by the
// LemonSqueezy API, which uses pointers where the webhook payload uses nulls.
func subscriptionSyncFromLemonSqueezy(
	lemonSqueezySubscriptionID string,
	userID string,
	appID string,
	sub lemonsqueezy.Subscription,
) subscriptionSync {
	return subscriptionSync{
		LemonSqueezySubscriptionID: lemonSqueezySubscriptionID,
		UserID:                     userID,
		AppID:                      appID,
		ProductName:                sub.ProductName,
		Status:                     sub.Status,
		StatusFormatted:            sub.StatusFormatted,
		LemonSqueezyCustomerID:     strconv.Itoa(sub.CustomerID),
		LemonSqueezyOrderID:        strconv.Itoa(sub.OrderID),
		LemonSqueezyProductID:      strconv.Itoa(sub.ProductID),
		LemonSqueezyVariantID:      strconv.Itoa(sub.VariantID),
		// The client decodes renews_at into a plain time.Time, so a null from
		// LemonSqueezy arrives as the zero value rather than as an absent date.
		RenewsAt:                   null.NewTime(sub.RenewsAt, !sub.RenewsAt.IsZero()),
		TrialEndsAt:                null.TimeFromPtr(sub.TrialEndsAt),
		EndsAt:                     null.TimeFromPtr(sub.EndsAt),
		CreatedAt:                  sub.CreatedAt,
		UpdatedAt:                  sub.UpdatedAt,
	}
}

// syncSubscription stores the subscription and brings its entitlements in line
// with the plan and dates LemonSqueezy reports.
func (h *BillingHandler) syncSubscription(ctx context.Context, in subscriptionSync) (*model.Subscription, error) {
	subscription, err := h.subscriptionStore.UpsertLemonSqueezySubscription(ctx, model.Subscription{
		ID:                         util.UniqueID(),
		DisplayName:                in.ProductName,
		Source:                     model.SubscriptionSourceLemonSqueezy,
		Status:                     in.Status,
		StatusFormatted:            in.StatusFormatted,
		RenewsAt:                   in.RenewsAt,
		TrialEndsAt:                in.TrialEndsAt,
		EndsAt:                     in.EndsAt,
		CreatedAt:                  in.CreatedAt,
		UpdatedAt:                  in.UpdatedAt,
		UserID:                     in.UserID,
		LemonsqueezySubscriptionID: null.StringFrom(in.LemonSqueezySubscriptionID),
		LemonsqueezyCustomerID:     null.StringFrom(in.LemonSqueezyCustomerID),
		LemonsqueezyOrderID:        null.StringFrom(in.LemonSqueezyOrderID),
		LemonsqueezyProductID:      null.StringFrom(in.LemonSqueezyProductID),
		LemonsqueezyVariantID:      null.StringFrom(in.LemonSqueezyVariantID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert subscription: %w", err)
	}

	plan := h.planManager.PlanByLemonSqueezyProductID(in.LemonSqueezyProductID)
	if plan == nil {
		return nil, fmt.Errorf("failed to find plan for product ID %s", in.LemonSqueezyProductID)
	}

	entitlement := model.Entitlement{
		ID:             util.UniqueID(),
		Type:           "subscription",
		SubscriptionID: null.StringFrom(subscription.ID),
		AppID:          in.AppID,
		PlanID:         plan.ID,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		EndsAt:         entitlementEndsAt(in.RenewsAt, in.EndsAt),
	}

	if in.AppID != "" {
		// Create a new entitlement or update the existing one
		if _, err := h.entitlementStore.UpsertSubscriptionEntitlement(ctx, entitlement); err != nil {
			return nil, fmt.Errorf("failed to upsert subscription entitlement: %w", err)
		}
	} else {
		// We don't have the app ID, but there might be an entitlement anyway, so we update that
		if err := h.entitlementStore.UpdateSubscriptionEntitlement(ctx, entitlement); err != nil {
			return nil, fmt.Errorf("failed to update subscription entitlement: %w", err)
		}
	}

	return subscription, nil
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
