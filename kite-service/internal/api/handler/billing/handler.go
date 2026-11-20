package billing

import (
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/core/billing"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type BillingHandlerConfig struct {
	LemonSqueezyStoreID       string
	LemonSqueezySigningSecret string
	TestMode                  bool
	AppPublicBaseURL          string
	Sale                      config.BillingSaleConfig
}

type BillingHandler struct {
	config              BillingHandlerConfig
	userStore           store.UserStore
	subscriptionStore   store.SubscriptionStore
	planManager         *plan.PlanManager
	subscriptionManager *billing.SubscriptionManager

	client *lemonsqueezy.Client
	now    func() time.Time
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
		now:    time.Now,
	}
}
