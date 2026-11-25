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
	TestMode                  bool
	AppPublicBaseURL          string
	Sale                      config.BillingSaleConfig
	Paddle                    config.BillingPaddleConfig

	// CheckoutProvider is where new subscriptions are bought.
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
	// paddleVerifier is nil without a Paddle webhook secret.
	paddleVerifier *paddle.WebhookVerifier
	now            func() time.Time
}

func NewBillingHandler(
	config BillingHandlerConfig,
	userStore store.UserStore,
	subscriptionStore store.SubscriptionStore,
	planManager *plan.PlanManager,
	subscriptionManager *billing.SubscriptionManager,
	client *lemonsqueezy.Client,
	paddleClient *paddle.SDK,
) *BillingHandler {
	var paddleVerifier *paddle.WebhookVerifier
	if config.Paddle.WebhookSecret != "" {
		paddleVerifier = paddle.NewWebhookVerifier(
			config.Paddle.WebhookSecret,
			paddle.VerifierWithTimestampTolerance(paddleWebhookTolerance),
		)
	}

	return &BillingHandler{
		config:              config,
		userStore:           userStore,
		subscriptionStore:   subscriptionStore,
		planManager:         planManager,
		subscriptionManager: subscriptionManager,

		client:         client,
		paddle:         paddleClient,
		paddleVerifier: paddleVerifier,
		now:            time.Now,
	}
}
