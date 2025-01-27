package wire

import (
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type BillingWebhookRequest struct {
	Meta struct {
		EventName  string                 `json:"event_name"`
		CustomData map[string]interface{} `json:"custom_data"`
	} `json:"meta"`
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			StoreID         int       `json:"store_id"`
			CustomerID      int       `json:"customer_id"`
			OrderID         int       `json:"order_id"`
			OrderItemID     int       `json:"order_item_id"`
			ProductID       int       `json:"product_id"`
			VariantID       int       `json:"variant_id"`
			ProductName     string    `json:"product_name"`
			VariantName     string    `json:"variant_name"`
			UserName        string    `json:"user_name"`
			UserEmail       string    `json:"user_email"`
			Status          string    `json:"status"`
			StatusFormatted string    `json:"status_formatted"`
			CardBrand       string    `json:"card_brand"`
			CardLastFour    string    `json:"card_last_four"`
			Cancelled       bool      `json:"cancelled"`
			TrialEndsAt     null.Time `json:"trial_ends_at"`
			BillingAnchor   int       `json:"billing_anchor"`
			RenewsAt        time.Time `json:"renews_at"`
			EndsAt          null.Time `json:"ends_at"`
			CreatedAt       time.Time `json:"created_at"`
			UpdatedAt       time.Time `json:"updated_at"`
			TestMode        bool      `json:"test_mode"`
		} `json:"attributes"`
	} `json:"data"`
}

type BillingWebhookResponse struct{}

type BillingCheckoutRequest struct{}

type BillingCheckoutResponse struct {
	URL string `json:"url"`
}

type BillingSubscription struct {
	ID string `json:"id"`
}

func SubscriptionToWire(subscription *model.Subscription) *BillingSubscription {
	if subscription == nil {
		return nil
	}

	return &BillingSubscription{
		ID: subscription.ID,
	}
}
