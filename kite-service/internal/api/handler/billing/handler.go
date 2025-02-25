package billing

import (
	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type BillingHandlerConfig struct {
	LemonSqueezyAPIKey        string
	LemonSqueezySigningSecret string
	LemonSqueezyStoreID       string
	TestMode                  bool
	AppPublicBaseURL          string
}

type BillingHandler struct {
	config            BillingHandlerConfig
	userStore         store.UserStore
	subscriptionStore store.SubscriptionStore
	entitlementStore  store.EntitlementStore
	planManager       *plan.PlanManager

	client *lemonsqueezy.Client
}

func NewBillingHandler(
	config BillingHandlerConfig,
	userStore store.UserStore,
	subscriptionStore store.SubscriptionStore,
	entitlementStore store.EntitlementStore,
	planManager *plan.PlanManager,
) *BillingHandler {
	client := lemonsqueezy.New(
		lemonsqueezy.WithAPIKey(config.LemonSqueezyAPIKey),
		lemonsqueezy.WithSigningSecret(config.LemonSqueezySigningSecret),
	)

	return &BillingHandler{
		config:            config,
		userStore:         userStore,
		subscriptionStore: subscriptionStore,
		entitlementStore:  entitlementStore,
		planManager:       planManager,

		client: client,
	}
}
