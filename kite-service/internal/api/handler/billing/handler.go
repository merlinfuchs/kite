package billing

import (
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/core/billing"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type BillingHandlerConfig struct {
	LemonSqueezyStoreID       string
	LemonSqueezySigningSecret string
	PaddleWebhookSecret       string
	PaddleClientToken         string
	PaddleSandbox             bool
	TestMode                  bool
	AppPublicBaseURL          string
	Sale                      config.BillingSaleConfig

	// CheckoutProvider is where new subscriptions are bought, Lemon Squeezy
	// if it's empty.
	CheckoutProvider model.SubscriptionSource
}

type BillingHandler struct {
	config              BillingHandlerConfig
	userStore           store.UserStore
	subscriptionStore   store.SubscriptionStore
	planManager         *plan.PlanManager
	subscriptionManager *billing.SubscriptionManager

	client *lemonsqueezy.Client
	// paddle is nil without a Paddle API key.
	paddle *paddle.SDK
	now    func() time.Time
}

func NewBillingHandler(
	config BillingHandlerConfig,
	userStore store.UserStore,
	subscriptionStore store.SubscriptionStore,
	planManager *plan.PlanManager,
	subscriptionManager *billing.SubscriptionManager,
	client *lemonsqueezy.Client,
	paddle *paddle.SDK,
) *BillingHandler {
	return &BillingHandler{
		config:              config,
		userStore:           userStore,
		subscriptionStore:   subscriptionStore,
		planManager:         planManager,
		subscriptionManager: subscriptionManager,

		client: client,
		paddle: paddle,
		now:    time.Now,
	}
}

func (h *BillingHandler) checkoutProvider() model.SubscriptionSource {
	if h.config.CheckoutProvider == "" {
		return model.SubscriptionSourceLemonSqueezy
	}
	return h.config.CheckoutProvider
}
