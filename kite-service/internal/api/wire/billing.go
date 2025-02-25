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

type BillingCheckoutRequest struct {
	LemonSqueezyVariantID string `json:"lemonsqueezy_variant_id"`
}

type BillingCheckoutResponse struct {
	URL string `json:"url"`
}

type SubscriptionManageResponse struct {
	UpdatePaymentMethodURL string `json:"update_payment_method_url"`
	CustomerPortalURL      string `json:"customer_portal_url"`
}

type Subscription struct {
	ID                         string      `json:"id"`
	DisplayName                string      `json:"display_name"`
	Source                     string      `json:"source"`
	Status                     string      `json:"status"`
	StatusFormatted            string      `json:"status_formatted"`
	CreatedAt                  time.Time   `json:"created_at"`
	UpdatedAt                  time.Time   `json:"updated_at"`
	RenewsAt                   time.Time   `json:"renews_at"`
	TrialEndsAt                null.Time   `json:"trial_ends_at"`
	EndsAt                     null.Time   `json:"ends_at"`
	UserID                     string      `json:"user_id"`
	LemonsqueezySubscriptionID null.String `json:"lemonsqueezy_subscription_id"`
	LemonsqueezyCustomerID     null.String `json:"lemonsqueezy_customer_id"`
	LemonsqueezyOrderID        null.String `json:"lemonsqueezy_order_id"`
	LemonsqueezyProductID      null.String `json:"lemonsqueezy_product_id"`
	LemonsqueezyVariantID      null.String `json:"lemonsqueezy_variant_id"`
	Manageable                 bool        `json:"manageable"`
}

type SubscriptionListResponse = []*Subscription

func SubscriptionToWire(subscription *model.Subscription, userID string) *Subscription {
	if subscription == nil {
		return nil
	}

	return &Subscription{
		ID:                         subscription.ID,
		DisplayName:                subscription.DisplayName,
		Source:                     string(subscription.Source),
		Status:                     subscription.Status,
		StatusFormatted:            subscription.StatusFormatted,
		CreatedAt:                  subscription.CreatedAt,
		UpdatedAt:                  subscription.UpdatedAt,
		RenewsAt:                   subscription.RenewsAt,
		TrialEndsAt:                subscription.TrialEndsAt,
		EndsAt:                     subscription.EndsAt,
		UserID:                     subscription.UserID,
		LemonsqueezySubscriptionID: subscription.LemonsqueezySubscriptionID,
		LemonsqueezyCustomerID:     subscription.LemonsqueezyCustomerID,
		LemonsqueezyOrderID:        subscription.LemonsqueezyOrderID,
		LemonsqueezyProductID:      subscription.LemonsqueezyProductID,
		LemonsqueezyVariantID:      subscription.LemonsqueezyVariantID,
		Manageable:                 subscription.UserID == userID && subscription.LemonsqueezySubscriptionID.Valid,
	}
}

type BillingPlan struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`
	Default     bool    `json:"default"`
	Popular     bool    `json:"popular"`
	Hidden      bool    `json:"hidden"`

	LemonSqueezyProductID string `json:"lemonsqueezy_product_id"`
	LemonSqueezyVariantID string `json:"lemonsqueezy_variant_id"`

	DiscordRoleID string `json:"discord_role_id"`

	FeatureMaxCollaborators     int  `json:"feature_max_collaborators"`
	FeatureUsageCreditsPerMonth int  `json:"feature_usage_credits_per_month"`
	FeatureMaxGuilds            int  `json:"feature_max_guilds"`
	FeaturePrioritySupport      bool `json:"feature_priority_support"`
}

type BillingPlanListResponse = []*BillingPlan
