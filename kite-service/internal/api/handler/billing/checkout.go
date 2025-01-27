package billing

import (
	"fmt"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func (h *BillingHandler) HandleAppCheckout(c *handler.Context, req wire.BillingCheckoutRequest) (*wire.BillingCheckoutResponse, error) {
	user, err := h.userStore.User(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	redirectURL := fmt.Sprintf("%s/apps/%s/billing", h.config.AppPublicBaseURL, c.App.ID)

	res, _, err := h.client.Checkouts.Create(c.Context(), 148760, 672453, &lemonsqueezy.CheckoutCreateAttributes{
		TestMode: ptr(h.config.TestMode),
		CheckoutOptions: lemonsqueezy.CheckoutCreateOptions{
			Embed: ptr(true),
		},
		CheckoutData: lemonsqueezy.CheckoutCreateData{
			Name:  user.DisplayName,
			Email: user.Email,
			Custom: map[string]any{
				"user_id": c.Session.UserID,
				"app_id":  c.App.ID,
			},
		},
		ProductOptions: lemonsqueezy.CheckoutCreateProductOptions{
			RedirectURL: redirectURL,
		},
		ExpiresAt: ptr(time.Now().UTC().Add(time.Hour).Format(time.RFC3339)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout: %w", err)
	}

	return &wire.BillingCheckoutResponse{
		URL: res.Data.Attributes.URL,
	}, nil
}

func ptr[T any](v T) *T {
	return &v
}
