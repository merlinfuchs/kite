package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

var ErrPaddleNotConfigured = errors.New("paddle is not configured")

// ErrUnknownOwner is returned for Paddle subscriptions without a user in
// their custom data that we haven't stored before, like ones made in Paddle's
// dashboard.
var ErrUnknownOwner = errors.New("subscription has no user_id in its custom data and isn't stored")

// NewPaddleClient builds the client the subscription manager and the API
// handler share, nil without an API key.
func NewPaddleClient(cfg config.BillingPaddleConfig) (*paddle.SDK, error) {
	if cfg.APIKey == "" {
		return nil, nil
	}
	if cfg.Sandbox {
		return paddle.NewSandbox(cfg.APIKey)
	}
	return paddle.New(cfg.APIKey)
}

// RefreshPaddle syncs the subscription as Paddle currently reports it. Its
// user and app are in the custom data our checkout sets, which Paddle keeps on
// the subscription.
func (m *SubscriptionManager) RefreshPaddle(ctx context.Context, paddleID string) (*model.Subscription, error) {
	res, err := m.getPaddleSubscription(ctx, paddleID)
	if err != nil {
		return nil, err
	}

	userID, _ := res.CustomData["user_id"].(string)
	appID, _ := res.CustomData["app_id"].(string)
	if userID == "" {
		// Subscriptions that weren't bought through our checkout, like ones
		// imported from Lemon Squeezy, have no custom data. Once we have
		// stored one, we know its owner.
		existing, err := m.subscriptionStore.SubscriptionByPaddleID(ctx, paddleID)
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrUnknownOwner
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get subscription without user_id in its custom data: %w", err)
		}
		userID = existing.UserID
	}

	if movePending(res) {
		if res, err = m.completeMove(ctx, res, userID, appID); err != nil {
			return nil, err
		}
	}

	sub, err := SubscriptionFromPaddle(res, userID)
	if err != nil {
		return nil, err
	}
	return m.Sync(ctx, sub, appID)
}

func (m *SubscriptionManager) changePaddlePlan(
	ctx context.Context,
	subscription *model.Subscription,
	priceID string,
	appID string,
) (*model.Subscription, error) {
	if m.paddle == nil {
		return nil, ErrPaddleNotConfigured
	}

	// Replacing the items swaps the plan, also between monthly and yearly.
	res, err := m.paddle.UpdateSubscription(ctx, &paddle.UpdateSubscriptionRequest{
		SubscriptionID: subscription.PaddleSubscriptionID.String,
		Items: paddle.NewPatchField([]paddle.UpdateSubscriptionItems{{
			SubscriptionUpdateItemFromCatalog: &paddle.SubscriptionUpdateItemFromCatalog{PriceID: priceID, Quantity: 1},
		}}),
		ProrationBillingMode: paddle.NewPatchField(paddle.ProrationBillingModeProratedImmediately),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription in Paddle: %w", err)
	}

	// Paddle also sends a subscription.updated webhook, but we cannot rely on
	// it having arrived by the time the client refetches.
	sub, err := SubscriptionFromPaddle(res, subscription.UserID)
	if err != nil {
		return nil, err
	}
	return m.Sync(ctx, sub, appID)
}

func (m *SubscriptionManager) getPaddleSubscription(ctx context.Context, paddleID string) (*paddle.Subscription, error) {
	if m.paddle == nil {
		return nil, ErrPaddleNotConfigured
	}

	res, err := m.paddle.GetSubscription(ctx, &paddle.GetSubscriptionRequest{SubscriptionID: paddleID})
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription from Paddle: %w", err)
	}
	return res, nil
}

// SubscriptionFromPaddle adapts a subscription as returned by the Paddle API
// to the statuses of Lemon Squeezy, which the rest of Kite knows.
//
// A subscription that's cancelled or paused at the end of its period stays
// active in Paddle until then, with a scheduled change. It ends at that
// change. Paddle's canceled is final, like expired in Lemon Squeezy.
func SubscriptionFromPaddle(sub *paddle.Subscription, userID string) (model.Subscription, error) {
	if len(sub.Items) == 0 {
		return model.Subscription{}, fmt.Errorf("paddle subscription %s has no items", sub.ID)
	}
	// Checkout only sells a single plan, so its price is the only item.
	item := sub.Items[0]

	createdAt, err := parsePaddleTime(sub.CreatedAt)
	if err != nil {
		return model.Subscription{}, err
	}
	updatedAt, err := parsePaddleTime(sub.UpdatedAt)
	if err != nil {
		return model.Subscription{}, err
	}
	renewsAt, err := parseOptionalPaddleTime(sub.NextBilledAt)
	if err != nil {
		return model.Subscription{}, err
	}

	var endsAt, trialEndsAt null.Time
	switch {
	case sub.Status == paddle.SubscriptionStatusCanceled:
		endsAt, err = parseOptionalPaddleTime(sub.CanceledAt)
	case sub.ScheduledChange != nil && sub.ScheduledChange.Action != paddle.ScheduledChangeActionResume:
		endsAt, err = parseOptionalPaddleTime(&sub.ScheduledChange.EffectiveAt)
	case sub.Status == paddle.SubscriptionStatusTrialing && sub.CurrentBillingPeriod != nil:
		trialEndsAt, err = parseOptionalPaddleTime(&sub.CurrentBillingPeriod.EndsAt)
	}
	if err != nil {
		return model.Subscription{}, err
	}

	status, statusFormatted := paddleStatus(sub)

	return model.Subscription{
		UserID:               userID,
		Source:               model.SubscriptionSourcePaddle,
		DisplayName:          item.Product.Name,
		Status:               status,
		StatusFormatted:      statusFormatted,
		RenewsAt:             renewsAt,
		TrialEndsAt:          trialEndsAt,
		EndsAt:               endsAt,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
		PaddleSubscriptionID: null.StringFrom(sub.ID),
		PaddleCustomerID:     null.StringFrom(sub.CustomerID),
		PaddleProductID:      null.StringFrom(item.Price.ProductID),
		PaddlePriceID:        null.StringFrom(item.Price.ID),
	}, nil
}

// paddleStatus is the Lemon Squeezy status of the Paddle subscription and how
// Lemon Squeezy writes it.
func paddleStatus(sub *paddle.Subscription) (string, string) {
	switch sub.Status {
	case paddle.SubscriptionStatusTrialing:
		return model.SubscriptionStatusOnTrial, "On Trial"
	case paddle.SubscriptionStatusPastDue:
		return model.SubscriptionStatusPastDue, "Past due"
	case paddle.SubscriptionStatusPaused:
		return model.SubscriptionStatusPaused, "Paused"
	case paddle.SubscriptionStatusCanceled:
		return model.SubscriptionStatusExpired, "Expired"
	case paddle.SubscriptionStatusActive:
		if sub.ScheduledChange != nil && sub.ScheduledChange.Action == paddle.ScheduledChangeActionCancel {
			return model.SubscriptionStatusCancelled, "Cancelled"
		}
		return model.SubscriptionStatusActive, "Active"
	}
	// A status Paddle adds later doesn't give access until we know it.
	return string(sub.Status), string(sub.Status)
}

func parsePaddleTime(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse Paddle time %q: %w", value, err)
	}
	return t.UTC(), nil
}

func parseOptionalPaddleTime(value *string) (null.Time, error) {
	if value == nil || *value == "" {
		return null.Time{}, nil
	}
	t, err := parsePaddleTime(*value)
	if err != nil {
		return null.Time{}, err
	}
	return null.TimeFrom(t), nil
}
