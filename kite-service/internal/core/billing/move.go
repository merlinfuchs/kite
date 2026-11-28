package billing

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// MoveLead is how long before it renews a Lemon Squeezy subscription can
// still move to Paddle. Paddle doesn't move a billing date to less than 30
// minutes away, and the customer needs time to check out.
const MoveLead = time.Hour

// The custom data of a checkout that moves a Lemon Squeezy subscription, next
// to the user_id and app_id of every checkout.
const (
	// moveFromKey is the ID of the Lemon Squeezy subscription.
	moveFromKey = "moved_from_lemonsqueezy_id"
	// movePriceKey is the Paddle price that replaces the $0 price of the
	// checkout.
	movePriceKey = "move_price_id"
	// movedKey is set once the move is complete.
	movedKey = "moved"
)

// ErrInvalidMove is returned for a move from Lemon Squeezy that can't be
// completed, like one of another user's subscription.
var ErrInvalidMove = errors.New("invalid move from Lemon Squeezy")

// MoveCustomData is the custom data of the checkout that moves subscription
// to the Paddle price priceID.
func MoveCustomData(subscription *model.Subscription, appID string, priceID string) paddle.CustomData {
	return paddle.CustomData{
		"user_id":    subscription.UserID,
		"app_id":     appID,
		moveFromKey:  subscription.LemonsqueezySubscriptionID.String,
		movePriceKey: priceID,
	}
}

func movePending(sub *paddle.Subscription) bool {
	from, _ := sub.CustomData[moveFromKey].(string)
	moved, _ := sub.CustomData[movedKey].(bool)
	return from != "" && !moved
}

// completeMove finishes moving a Lemon Squeezy subscription to the Paddle
// subscription its checkout created. The checkout was for $0, so the customer
// hasn't paid anything yet. This switches to the plan's price without billing
// it, cancels the Lemon Squeezy subscription, which stays active until the end
// of its paid period, and moves the first payment on Paddle to that date.
//
// Paddle can send the webhooks of a subscription more than once, so each step
// is skipped once done.
func (m *SubscriptionManager) completeMove(
	ctx context.Context,
	sub *paddle.Subscription,
	userID string,
	appID string,
) (*paddle.Subscription, error) {
	lemonSqueezyID, _ := sub.CustomData[moveFromKey].(string)
	priceID, _ := sub.CustomData[movePriceKey].(string)
	if priceID == "" {
		return nil, fmt.Errorf("%w: paddle subscription %s has no price to move to", ErrInvalidMove, sub.ID)
	}
	if m.client == nil {
		return nil, errors.New("can't move a Lemon Squeezy subscription without its API key")
	}

	stored, err := m.subscriptionStore.SubscriptionByLemonSqueezyID(ctx, lemonSqueezyID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("%w: lemon squeezy subscription %s isn't stored", ErrInvalidMove, lemonSqueezyID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get the Lemon Squeezy subscription to move: %w", err)
	}
	if stored.UserID != userID {
		return nil, fmt.Errorf("%w: lemon squeezy subscription %s belongs to another user than Paddle subscription %s", ErrInvalidMove, lemonSqueezyID, sub.ID)
	}

	if sub.Items[0].Price.ID != priceID {
		sub, err = m.paddle.UpdateSubscription(ctx, &paddle.UpdateSubscriptionRequest{
			SubscriptionID: sub.ID,
			Items: paddle.NewPatchField([]paddle.UpdateSubscriptionItems{{
				SubscriptionUpdateItemFromCatalog: &paddle.SubscriptionUpdateItemFromCatalog{PriceID: priceID, Quantity: 1},
			}}),
			ProrationBillingMode: paddle.NewPatchField(paddle.ProrationBillingModeDoNotBill),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to switch moved subscription to its price in Paddle: %w", err)
		}
	}

	res, _, err := m.client.Subscriptions.Get(ctx, lemonSqueezyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription from LemonSqueezy: %w", err)
	}
	if status := res.Data.Attributes.Status; status != model.SubscriptionStatusCancelled && status != model.SubscriptionStatusExpired {
		res, _, err = m.client.Subscriptions.Cancel(ctx, lemonSqueezyID)
		if err != nil {
			return nil, fmt.Errorf("failed to cancel moved subscription in LemonSqueezy: %w", err)
		}
	}
	lemonSqueezySub := SubscriptionFromLemonSqueezy(res.Data.ID, res.Data.Attributes, userID)
	if _, err := m.Sync(ctx, lemonSqueezySub, appID); err != nil {
		return nil, fmt.Errorf("failed to sync moved Lemon Squeezy subscription: %w", err)
	}

	// The paid period on Lemon Squeezy can have ended by now, if the webhook
	// took that long to arrive.
	firstBilledAt := time.Now().UTC().Add(MoveLead)
	if paidUntil := lemonSqueezySub.PaidUntil(); paidUntil.Valid && paidUntil.Time.After(firstBilledAt) {
		firstBilledAt = paidUntil.Time
	}

	customData := maps.Clone(sub.CustomData)
	customData[movedKey] = true
	sub, err = m.paddle.UpdateSubscription(ctx, &paddle.UpdateSubscriptionRequest{
		SubscriptionID:       sub.ID,
		NextBilledAt:         paddle.NewPatchField(firstBilledAt.Format(time.RFC3339)),
		CustomData:           paddle.NewPatchField(customData),
		ProrationBillingMode: paddle.NewPatchField(paddle.ProrationBillingModeDoNotBill),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to move the billing date of moved subscription in Paddle: %w", err)
	}
	return sub, nil
}
