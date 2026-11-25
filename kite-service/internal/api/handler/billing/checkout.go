package billing

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func (h *BillingHandler) HandleAppCheckout(c *handler.Context, req wire.BillingCheckoutRequest) (*wire.BillingCheckoutResponse, error) {
	provider := h.checkoutProvider()
	yearly := req.Interval == wire.BillingIntervalYearly

	plan := h.planManager.PlanByID(req.PlanID)
	if plan == nil || plan.Hidden {
		return nil, handler.ErrBadRequest("unknown_plan", "There is no such plan")
	}
	priceID := plan.PriceID(provider, yearly)
	if priceID == "" {
		return nil, handler.ErrBadRequest("unknown_plan", fmt.Sprintf("The plan can't be bought %s", req.Interval))
	}

	user, err := h.userStore.User(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	redirectURL := fmt.Sprintf("%s/apps/%s/premium", h.config.AppPublicBaseURL, c.App.ID)

	if provider == model.SubscriptionSourcePaddle {
		return h.paddleCheckout(c, user, priceID, yearly, redirectURL)
	}

	variantID, err := strconv.Atoi(priceID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert variant ID to int: %w", err)
	}

	storeID, err := strconv.Atoi(h.config.LemonSqueezyStoreID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert store ID to int: %w", err)
	}

	var discountCode string
	if yearly && h.saleActive() {
		discountCode = h.config.Sale.Code
	}

	res, _, err := h.client.Checkouts.Create(
		c.Context(),
		storeID,
		variantID,
		&lemonsqueezy.CheckoutCreateAttributes{
			TestMode: ptr(h.config.TestMode),
			CheckoutOptions: lemonsqueezy.CheckoutCreateOptions{
				Embed: ptr(true),
			},
			CheckoutData: lemonsqueezy.CheckoutCreateData{
				Name:         user.DisplayName,
				Email:        user.Email,
				DiscountCode: discountCode,
				Custom: map[string]any{
					"user_id": c.Session.UserID,
					"app_id":  c.App.ID,
				},
			},
			ProductOptions: lemonsqueezy.CheckoutCreateProductOptions{
				RedirectURL: redirectURL,
			},
			ExpiresAt: ptr(time.Now().UTC().Add(time.Hour).Format(time.RFC3339)),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout: %w", err)
	}

	return &wire.BillingCheckoutResponse{
		Provider: string(model.SubscriptionSourceLemonSqueezy),
		URL:      res.Data.Attributes.URL,
	}, nil
}

// paddleCheckout creates the transaction that Paddle.js opens the checkout of.
// The user and app are set here rather than in the browser, so they can't be
// changed, and Paddle keeps them on the subscription.
func (h *BillingHandler) paddleCheckout(
	c *handler.Context,
	user *model.User,
	priceID string,
	yearly bool,
	successURL string,
) (*wire.BillingCheckoutResponse, error) {
	if h.paddle == nil {
		return nil, fmt.Errorf("checkout goes to Paddle, but Paddle is not configured")
	}

	req := &paddle.CreateTransactionRequest{
		Items: []paddle.CreateTransactionItems{
			*paddle.NewCreateTransactionItemsTransactionItemFromCatalog(&paddle.TransactionItemFromCatalog{
				PriceID:  priceID,
				Quantity: 1,
			}),
		},
		CustomData: paddle.CustomData{
			"user_id": c.Session.UserID,
			"app_id":  c.App.ID,
		},
	}
	if yearly && h.saleActive() {
		req.DiscountID = ptr(h.config.Sale.PaddleDiscountID)
	}

	txn, err := h.paddle.CreateTransaction(c.Context(), req)
	if err != nil {
		slog.Error(
			"Failed to create Paddle transaction",
			slog.String("price_id", priceID),
			slog.String("error", err.Error()),
		)
		return nil, handler.ErrInternal("Failed to reach our payment provider, please try again")
	}

	return &wire.BillingCheckoutResponse{
		Provider:            string(model.SubscriptionSourcePaddle),
		PaddleTransactionID: txn.ID,
		PaddleClientToken:   h.config.PaddleClientToken,
		PaddleSandbox:       h.config.PaddleSandbox,
		CustomerEmail:       user.Email,
		SuccessURL:          successURL,
	}, nil
}

func ptr[T any](v T) *T {
	return &v
}
