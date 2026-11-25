package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/config"
	corebilling "github.com/kitecloud/kite/kite-service/internal/core/billing"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type fakeUserStore struct {
	store.UserStore
}

func (fakeUserStore) User(ctx context.Context, id string) (*model.User, error) {
	return &model.User{ID: id, Email: "user@example.com"}, nil
}

var paddlePlans = []model.Plan{
	{ID: "basic"},
	{
		ID:                          "plus",
		LemonSqueezyProductID:       "1",
		LemonSqueezyVariantID:       "10",
		LemonSqueezyYearlyVariantID: "11",
		PaddleProductID:             "pro_plus",
		PaddlePriceID:               "pri_plus",
		PaddleYearlyPriceID:         "pri_plus_yearly",
	},
	// Only sold monthly through Paddle.
	{ID: "premium", PaddleProductID: "pro_premium", PaddlePriceID: "pri_premium", LemonSqueezyYearlyVariantID: "21"},
}

type testBillingServer struct {
	http.Handler
	subStore *fakeSubscriptionStore
	// requests are the method, path and body of each request to the Paddle API.
	requests []string
	// response is what the stubbed Paddle API answers.
	response string
}

func newTestBillingServer(t *testing.T, cfg BillingHandlerConfig) *testBillingServer {
	s := &testBillingServer{subStore: &fakeSubscriptionStore{}}

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.requests = append(s.requests, r.Method+" "+r.URL.Path+" "+string(bytes.TrimSpace(body)))
		_, _ = w.Write([]byte(`{"data": ` + s.response + `}`))
	}))
	t.Cleanup(api.Close)

	entStore := &fakeEntitlementStore{}
	planManager := plan.NewPlanManager(entStore, s.subStore, nil, nil, paddlePlans, plan.PlanManagerConfig{})
	paddleClient, err := paddle.New("key", paddle.WithBaseURL(api.URL))
	require.NoError(t, err)
	// Both providers' APIs are stubbed by the same server.
	lsClient := lemonsqueezy.New(lemonsqueezy.WithBaseURL(api.URL), lemonsqueezy.WithAPIKey("key"))
	subManager := corebilling.NewSubscriptionManager(s.subStore, entStore, planManager, lsClient, paddleClient)

	cfg.AppPublicBaseURL = "https://kite.onl"
	cfg.LemonSqueezyStoreID = "1"
	h := NewBillingHandler(cfg, fakeUserStore{}, s.subStore, planManager, subManager, lsClient, paddleClient)

	// Stands in for the session and app middleware.
	withApp := func(next handler.HandlerFunc) handler.HandlerFunc {
		return func(c *handler.Context) error {
			c.Session = &model.Session{UserID: "user-1"}
			c.App = &model.App{ID: "app-1"}
			return next(c)
		}
	}
	mux := http.NewServeMux()
	group := handler.Group(mux, "/v1")
	group.Get("/billing/plans", handler.Typed(h.HandleBillingPlanList))
	group.Post("/checkout", withApp(handler.TypedWithBody(h.HandleAppCheckout)))
	group.Put("/subscriptions/{subscriptionID}/plan", withApp(handler.TypedWithBody(h.HandleSubscriptionPlanUpdate)))
	s.Handler = mux
	return s
}

func request[T any](t *testing.T, server http.Handler, method string, path string, body string) (int, wire.APIResponse[T]) {
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	var resp wire.APIResponse[T]
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp), rec.Body.String())
	return rec.Code, resp
}

func TestPaddleCheckout(t *testing.T) {
	cfg := BillingHandlerConfig{
		CheckoutProvider: model.SubscriptionSourcePaddle,
		Paddle:           config.BillingPaddleConfig{ClientToken: "test_token", Sandbox: true},
	}

	t.Run("creates a transaction with the user and app", func(t *testing.T) {
		server := newTestBillingServer(t, cfg)
		server.response = `{"id": "txn_1"}`

		code, resp := request[wire.BillingCheckoutResponse](t, server, http.MethodPost, "/v1/checkout", `{"plan_id": "plus", "interval": "yearly"}`)
		require.Equal(t, http.StatusOK, code)
		assert.Equal(t, wire.BillingCheckoutResponse{
			Provider:            "paddle",
			PaddleTransactionID: "txn_1",
			PaddleClientToken:   "test_token",
			PaddleSandbox:       true,
			CustomerEmail:       "user@example.com",
			SuccessURL:          "https://kite.onl/apps/app-1/premium",
		}, resp.Data)

		require.Len(t, server.requests, 1)
		assert.Equal(t, `POST /transactions {"items":[{"quantity":1,"price_id":"pri_plus_yearly"}],"custom_data":{"app_id":"app-1","user_id":"user-1"}}`, server.requests[0])
	})

	t.Run("applies the sale to yearly plans", func(t *testing.T) {
		cfg := cfg
		cfg.Sale = config.BillingSaleConfig{
			Name: "Black Week", Code: "BW", Percent: 30, PaddleDiscountID: "dsc_1",
			StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(time.Hour),
		}
		server := newTestBillingServer(t, cfg)
		server.response = `{"id": "txn_1"}`

		request[wire.BillingCheckoutResponse](t, server, http.MethodPost, "/v1/checkout", `{"plan_id": "plus", "interval": "yearly"}`)
		request[wire.BillingCheckoutResponse](t, server, http.MethodPost, "/v1/checkout", `{"plan_id": "plus", "interval": "monthly"}`)
		require.Len(t, server.requests, 2)
		assert.Contains(t, server.requests[0], `"discount_id":"dsc_1"`)
		assert.NotContains(t, server.requests[1], `discount_id`)
	})

	t.Run("rejects plans Paddle doesn't sell", func(t *testing.T) {
		server := newTestBillingServer(t, cfg)

		for _, body := range []string{
			`{"plan_id": "basic", "interval": "monthly"}`,
			`{"plan_id": "premium", "interval": "yearly"}`,
			`{"plan_id": "unknown", "interval": "monthly"}`,
			`{"plan_id": "plus", "interval": "weekly"}`,
		} {
			code, resp := request[wire.BillingCheckoutResponse](t, server, http.MethodPost, "/v1/checkout", body)
			assert.Equal(t, http.StatusBadRequest, code, body)
			assert.False(t, resp.Success)
		}
		assert.Empty(t, server.requests)
	})
}

func TestLemonSqueezyCheckoutSale(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)
	for name, tt := range map[string]struct {
		interval string
		start    time.Time
		discount bool
	}{
		"yearly during the sale":  {"yearly", start, true},
		"monthly during the sale": {"monthly", start, false},
		"yearly before the sale":  {"yearly", end, false},
	} {
		t.Run(name, func(t *testing.T) {
			server := newTestBillingServer(t, BillingHandlerConfig{
				CheckoutProvider: model.SubscriptionSourceLemonSqueezy,
				Sale:             config.BillingSaleConfig{Name: "Black Week", Code: "BW", Percent: 30, StartsAt: tt.start, EndsAt: end},
			})
			server.response = `{"id": "1", "attributes": {"url": "https://kite.lemonsqueezy.com/checkout/1"}}`

			code, resp := request[wire.BillingCheckoutResponse](t, server, http.MethodPost, "/v1/checkout", `{"plan_id": "plus", "interval": "`+tt.interval+`"}`)
			require.Equal(t, http.StatusOK, code, resp.Error)
			assert.Equal(t, "https://kite.lemonsqueezy.com/checkout/1", resp.Data.URL)
			require.Len(t, server.requests, 1)
			assert.Equal(t, tt.discount, strings.Contains(server.requests[0], `"discount_code":"BW"`), server.requests[0])
		})
	}
}

func TestBillingPlanListYearlyPrice(t *testing.T) {
	plans := []model.Plan{
		{ID: "plus", YearlyPrice: 29.99, LemonSqueezyYearlyVariantID: "11", PaddleYearlyPriceID: "pri_plus_yearly"},
		{ID: "premium", YearlyPrice: 59.99, LemonSqueezyYearlyVariantID: "21"},
	}

	for provider, want := range map[model.SubscriptionSource][]float32{
		model.SubscriptionSourceLemonSqueezy: {29.99, 59.99},
		model.SubscriptionSourcePaddle:       {29.99, 0},
	} {
		h := NewBillingHandler(BillingHandlerConfig{CheckoutProvider: provider}, nil, nil,
			plan.NewPlanManager(nil, nil, nil, nil, plans, plan.PlanManagerConfig{}), nil, nil, nil)
		res, err := h.HandleBillingPlanList(nil)
		require.NoError(t, err)
		assert.Equal(t, want, []float32{(*res)[0].YearlyPrice, (*res)[1].YearlyPrice}, provider)
	}
}

func TestPaddlePlanSwitch(t *testing.T) {
	// Checkout still goes to Lemon Squeezy, but the subscription is from Paddle.
	server := newTestBillingServer(t, BillingHandlerConfig{})
	server.subStore.byApp = []*model.Subscription{{
		ID:                   "sub-1",
		UserID:               "user-1",
		Source:               model.SubscriptionSourcePaddle,
		Status:               model.SubscriptionStatusActive,
		PaddleSubscriptionID: null.StringFrom("sub_1"),
		PaddleProductID:      null.StringFrom("pro_plus"),
		PaddlePriceID:        null.StringFrom("pri_plus"),
	}}
	server.response = `{
		"id": "sub_1",
		"status": "active",
		"customer_id": "ctm_1",
		"created_at": "2026-11-01T00:00:00Z",
		"updated_at": "2026-11-02T00:00:00Z",
		"next_billed_at": "2027-11-02T00:00:00Z",
		"items": [{"price": {"id": "pri_plus_yearly", "product_id": "pro_plus"}, "product": {"name": "Plus"}}],
		"custom_data": {"user_id": "user-1", "app_id": "app-1"}
	}`

	code, resp := request[wire.Subscription](t, server, http.MethodPut, "/v1/subscriptions/sub-1/plan", `{"plan_id": "plus", "interval": "yearly"}`)
	require.Equal(t, http.StatusOK, code, resp.Error)
	assert.Equal(t, "plus", resp.Data.PlanID)
	assert.Equal(t, wire.BillingIntervalYearly, resp.Data.Interval)

	require.Len(t, server.requests, 1)
	assert.Equal(t, `PATCH /subscriptions/sub_1 {"items":[{"price_id":"pri_plus_yearly","quantity":1}],"proration_billing_mode":"prorated_immediately"}`, server.requests[0])

	t.Run("rejects the plan it's on", func(t *testing.T) {
		server.requests = nil
		server.subStore.byApp[0].PaddlePriceID = null.StringFrom("pri_plus_yearly")

		code, resp := request[wire.Subscription](t, server, http.MethodPut, "/v1/subscriptions/sub-1/plan", `{"plan_id": "plus", "interval": "yearly"}`)
		assert.Equal(t, http.StatusBadRequest, code)
		assert.Equal(t, "plan_unchanged", resp.Error.Code)
		assert.Empty(t, server.requests)
	})
}
