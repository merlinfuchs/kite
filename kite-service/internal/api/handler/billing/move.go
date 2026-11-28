package billing

import (
	"fmt"

	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	corebilling "github.com/kitecloud/kite/kite-service/internal/core/billing"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// HandleSubscriptionMove opens a Paddle checkout that moves a Lemon Squeezy
// subscription to Paddle, before Lemon Squeezy shuts down. The checkout is for
// $0 and only saves the payment details. The webhook of the new subscription
// then puts it on the plan's price, billed from when the Lemon Squeezy
// subscription would renew, and cancels that one. See
// corebilling.SubscriptionManager.completeMove.
func (h *BillingHandler) HandleSubscriptionMove(c *handler.Context) (*wire.SubscriptionMoveResponse, error) {
	subscription, err := h.appSubscription(c)
	if err != nil {
		return nil, err
	}

	target := h.moveTarget(subscription, c.Session.UserID)
	if target == nil {
		return nil, handler.ErrBadRequest("unmovable_subscription", "This subscription can't be moved")
	}

	user, err := h.userStore.User(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	interval, amount := "month", target.plan.Price
	if target.yearly {
		interval, amount = "year", target.plan.YearlyPrice
	}

	req := &paddle.CreateTransactionRequest{
		Items: []paddle.CreateTransactionItems{
			*paddle.NewCreateTransactionItemsTransactionItemCreateWithPrice(&paddle.TransactionItemCreateWithPrice{
				Quantity: 1,
				Price: paddle.TransactionPriceCreateWithProductID{
					Description: "Moves a Lemon Squeezy subscription",
					// Shown in the checkout, which would otherwise only say
					// it costs $0.
					Name: ptr(fmt.Sprintf(
						"$0 today, then $%.2f a %s from %s",
						amount, interval, subscription.PaidUntil().Time.Format("January 2, 2006"),
					)),
					ProductID:    target.plan.PaddleProductID,
					BillingCycle: &paddle.Duration{Interval: paddle.Interval(interval), Frequency: 1},
					UnitPrice:    paddle.Money{Amount: "0", CurrencyCode: paddle.CurrencyCodeUSD},
				},
			}),
		},
		CustomData: corebilling.MoveCustomData(subscription, c.App.ID, target.priceID),
	}

	successURL := fmt.Sprintf("%s/apps/%s/premium", h.config.AppPublicBaseURL, c.App.ID)
	return h.openPaddleTransaction(c, user, req, successURL)
}

type moveTarget struct {
	plan    *model.Plan
	yearly  bool
	priceID string
}

// moveTarget is the Paddle price an active Lemon Squeezy subscription of the
// user moves to, the one of its plan and interval, nil if it can't move. It
// can't for a while before it renews, so it doesn't renew on Lemon Squeezy
// while the customer checks out. A subscription with a variant the plan no
// longer has, from before a price change, stays on Lemon Squeezy, because the
// price on Paddle would be different.
func (h *BillingHandler) moveTarget(subscription *model.Subscription, userID string) *moveTarget {
	if h.config.CheckoutProvider != model.SubscriptionSourcePaddle || h.paddle == nil {
		return nil
	}
	if subscription.Source != model.SubscriptionSourceLemonSqueezy ||
		subscription.UserID != userID ||
		subscription.Status != model.SubscriptionStatusActive {
		return nil
	}

	paidUntil := subscription.PaidUntil()
	if !paidUntil.Valid || paidUntil.Time.Before(h.now().Add(corebilling.MoveLead)) {
		return nil
	}

	plan := h.planManager.PlanBySubscription(subscription)
	if plan == nil {
		return nil
	}
	var yearly bool
	switch subscription.PriceID() {
	case plan.PriceID(model.SubscriptionSourceLemonSqueezy, false):
	case plan.PriceID(model.SubscriptionSourceLemonSqueezy, true):
		yearly = true
	default:
		return nil
	}

	priceID := plan.PriceID(model.SubscriptionSourcePaddle, yearly)
	if priceID == "" {
		return nil
	}
	return &moveTarget{plan: plan, yearly: yearly, priceID: priceID}
}
