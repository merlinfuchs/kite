package billing

import (
	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/core/billing"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type BillingHandlerConfig struct {
	LemonSqueezyStoreID string
	TestMode            bool
	AppPublicBaseURL    string
}

type BillingHandler struct {
	config              BillingHandlerConfig
	userStore           store.UserStore
	subscriptionStore   store.SubscriptionStore
	planManager         *plan.PlanManager
	subscriptionManager *billing.SubscriptionManager

	client *lemonsqueezy.Client
}

func NewBillingHandler(
	config BillingHandlerConfig,
	userStore store.UserStore,
	subscriptionStore store.SubscriptionStore,
	planManager *plan.PlanManager,
	subscriptionManager *billing.SubscriptionManager,
	client *lemonsqueezy.Client,
) *BillingHandler {
	return &BillingHandler{
		config:              config,
		userStore:           userStore,
		subscriptionStore:   subscriptionStore,
		planManager:         planManager,
		subscriptionManager: subscriptionManager,

		client: client,
	}
}
