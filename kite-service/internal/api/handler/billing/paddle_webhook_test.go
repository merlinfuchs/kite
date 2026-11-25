package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/config"
	corebilling "github.com/kitecloud/kite/kite-service/internal/core/billing"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type testPaddleWebhookServer struct {
	http.Handler
	subStore *fakeSubscriptionStore
	entStore *fakeEntitlementStore
	// customData is the custom data the stubbed Paddle API reports for the
	// subscription.
	customData string
}

func newTestPaddleWebhookServer(t *testing.T, secret string) *testPaddleWebhookServer {
	s := &testPaddleWebhookServer{
		subStore:   &fakeSubscriptionStore{},
		entStore:   &fakeEntitlementStore{},
		customData: `{"user_id": "user-456", "app_id": "app-123"}`,
	}

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions/sub_1" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"data": {
			"id": "sub_1",
			"status": "active",
			"customer_id": "ctm_1",
			"created_at": "2026-11-01T00:00:00Z",
			"updated_at": "2026-11-01T00:00:10Z",
			"next_billed_at": "2026-12-01T00:00:00Z",
			"items": [{"price": {"id": "pri_1", "product_id": "pro_1"}, "product": {"name": "Pro"}}],
			"custom_data": ` + s.customData + `
		}}`))
	}))
	t.Cleanup(api.Close)

	plans := []model.Plan{{ID: "pro", PaddleProductID: "pro_1"}}
	planManager := plan.NewPlanManager(s.entStore, s.subStore, nil, nil, plans, plan.PlanManagerConfig{})

	paddleClient, err := paddle.New("key", paddle.WithBaseURL(api.URL))
	require.NoError(t, err)
	lsClient := lemonsqueezy.New(lemonsqueezy.WithAPIKey("key"))
	subManager := corebilling.NewSubscriptionManager(s.subStore, s.entStore, planManager, lsClient, paddleClient)

	s.Handler = newTestWebhookServer(NewBillingHandler(
		BillingHandlerConfig{Paddle: config.BillingPaddleConfig{WebhookSecret: secret}},
		nil, s.subStore, planManager, subManager, lsClient,
	))
	return s
}

func paddleSignature(secret string, ts time.Time, payload []byte) string {
	tsStr := strconv.FormatInt(ts.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(tsStr + ":"))
	mac.Write(payload)
	return "ts=" + tsStr + ";h1=" + hex.EncodeToString(mac.Sum(nil))
}

func postPaddleWebhook(server http.Handler, signature string, payload []byte) *httptest.ResponseRecorder {
	return postWebhook(server, map[string]string{"Paddle-Signature": signature}, payload)
}

// A trailing newline is part of what Paddle signs, so the body must be
// verified as it came.
var paddleSubscriptionEvent = []byte(`{"event_id": "evt_1", "event_type": "subscription.created", "data": {"id": "sub_1"}}` + "\n")

func TestPaddleWebhookRejectsEmptySecret(t *testing.T) {
	server := newTestPaddleWebhookServer(t, "")

	rec := postPaddleWebhook(server, paddleSignature("", time.Now(), paddleSubscriptionEvent), paddleSubscriptionEvent)
	assertWebhookUnauthorized(t, rec)
	assert.Nil(t, server.subStore.upserted)
}

func TestPaddleWebhookRejectsInvalidSignature(t *testing.T) {
	server := newTestPaddleWebhookServer(t, testSecret)

	t.Run("signed with another secret", func(t *testing.T) {
		rec := postPaddleWebhook(server, paddleSignature("other", time.Now(), paddleSubscriptionEvent), paddleSubscriptionEvent)
		assertWebhookUnauthorized(t, rec)
	})

	t.Run("without a signature", func(t *testing.T) {
		rec := postPaddleWebhook(server, "", paddleSubscriptionEvent)
		assertWebhookUnauthorized(t, rec)
	})

	t.Run("signed too long ago", func(t *testing.T) {
		ts := time.Now().Add(-time.Hour)
		rec := postPaddleWebhook(server, paddleSignature(testSecret, ts, paddleSubscriptionEvent), paddleSubscriptionEvent)
		assertWebhookUnauthorized(t, rec)
	})

	assert.Nil(t, server.subStore.upserted)
}

func TestPaddleWebhookAcknowledgesUnrelatedEvents(t *testing.T) {
	server := newTestPaddleWebhookServer(t, testSecret)

	payload := []byte(`{"event_type": "transaction.completed", "data": {"id": "txn_1"}}`)
	rec := postPaddleWebhook(server, paddleSignature(testSecret, time.Now(), payload), payload)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Nil(t, server.subStore.upserted)
}

func TestPaddleWebhookSyncsSubscription(t *testing.T) {
	server := newTestPaddleWebhookServer(t, testSecret)

	rec := postPaddleWebhook(server, paddleSignature(testSecret, time.Now(), paddleSubscriptionEvent), paddleSubscriptionEvent)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	require.NotNil(t, server.subStore.upserted)
	assert.Equal(t, model.SubscriptionSourcePaddle, server.subStore.upserted.Source)
	assert.Equal(t, "user-456", server.subStore.upserted.UserID)
	assert.Equal(t, "sub_1", server.subStore.upserted.PaddleSubscriptionID.String)
	assert.Equal(t, "pri_1", server.subStore.upserted.PaddlePriceID.String)

	require.NotNil(t, server.entStore.upserted)
	assert.Equal(t, "app-123", server.entStore.upserted.AppID)
	assert.Equal(t, "pro", server.entStore.upserted.PlanID)
	assert.Equal(t, null.TimeFrom(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)), server.entStore.upserted.EndsAt)
}

func TestPaddleWebhookWithoutCustomData(t *testing.T) {
	t.Run("uses the owner of the stored subscription", func(t *testing.T) {
		server := newTestPaddleWebhookServer(t, testSecret)
		server.customData = `null`
		server.subStore.upserted = &model.Subscription{
			UserID:               "user-456",
			Source:               model.SubscriptionSourcePaddle,
			PaddleSubscriptionID: null.StringFrom("sub_1"),
		}

		rec := postPaddleWebhook(server, paddleSignature(testSecret, time.Now(), paddleSubscriptionEvent), paddleSubscriptionEvent)
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, "user-456", server.subStore.upserted.UserID)
		// Without an app, only the entitlements it has are updated
		assert.Nil(t, server.entStore.upserted)
		require.NotNil(t, server.entStore.updated)
	})

	t.Run("fails for an unknown subscription", func(t *testing.T) {
		server := newTestPaddleWebhookServer(t, testSecret)
		server.customData = `null`

		rec := postPaddleWebhook(server, paddleSignature(testSecret, time.Now(), paddleSubscriptionEvent), paddleSubscriptionEvent)
		assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
		assert.Nil(t, server.subStore.upserted)
	})
}
