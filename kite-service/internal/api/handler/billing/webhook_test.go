package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	corebilling "github.com/kitecloud/kite/kite-service/internal/core/billing"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type fakeSubscriptionStore struct {
	store.SubscriptionStore
	upserted *model.Subscription
}

func (s *fakeSubscriptionStore) UpsertLemonSqueezySubscription(ctx context.Context, sub model.Subscription) (*model.Subscription, error) {
	if s.upserted != nil && s.upserted.UpdatedAt.After(sub.UpdatedAt) {
		return nil, store.ErrOutdated
	}
	s.upserted = &sub
	return &sub, nil
}

func (s *fakeSubscriptionStore) SubscriptionByLemonSqueezyID(ctx context.Context, lemonSqueezyID string) (*model.Subscription, error) {
	if s.upserted != nil && s.upserted.LemonsqueezySubscriptionID.String == lemonSqueezyID {
		return s.upserted, nil
	}
	return nil, store.ErrNotFound
}

type fakeEntitlementStore struct {
	store.EntitlementStore
	upserted *model.Entitlement
}

func (s *fakeEntitlementStore) UpsertSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) (*model.Entitlement, error) {
	s.upserted = &entitlement
	return &entitlement, nil
}

func computeSignature(secret string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func newTestWebhookServer(h *BillingHandler) http.Handler {
	mux := http.NewServeMux()
	group := handler.Group(mux, "/v1")
	group.Post("/billing/webhook", handler.TypedWithBody(h.HandleBillingWebhook))
	return mux
}

func postWebhookRequest(server http.Handler, eventName string, signature string, payload []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/billing/webhook", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if eventName != "" {
		req.Header.Set("X-Event-Name", eventName)
	}
	if signature != "" {
		req.Header.Set("X-Signature", signature)
	}
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return rec
}

const testSecret = "valid-secret-key"

type testWebhookServer struct {
	http.Handler
	subStore *fakeSubscriptionStore
	entStore *fakeEntitlementStore
	// lsAttributes is what the stubbed LemonSqueezy API reports for the subscription.
	lsAttributes string
}

func newTestSubscriptionWebhookServer(t *testing.T, secret string) *testWebhookServer {
	s := &testWebhookServer{subStore: &fakeSubscriptionStore{}, entStore: &fakeEntitlementStore{}}

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":"ls-sub-789","attributes":` + s.lsAttributes + `}}`))
	}))
	t.Cleanup(api.Close)

	plans := []model.Plan{{ID: "pro", LemonSqueezyProductID: "1"}}
	planManager := plan.NewPlanManager(s.entStore, s.subStore, nil, plans, plan.PlanManagerConfig{})

	lsClient := lemonsqueezy.New(
		lemonsqueezy.WithBaseURL(api.URL),
		lemonsqueezy.WithAPIKey("key"),
		lemonsqueezy.WithSigningSecret(secret),
	)
	subManager := corebilling.NewSubscriptionManager(s.subStore, s.entStore, planManager, lsClient)

	s.Handler = newTestWebhookServer(NewBillingHandler(
		BillingHandlerConfig{LemonSqueezySigningSecret: secret},
		nil, s.subStore, planManager, subManager, lsClient,
	))
	return s
}

func subscriptionAttributes(productID string, updatedAt string) string {
	return `{
		"product_id": ` + productID + `,
		"status": "active",
		"created_at": "2026-01-01T00:00:00Z",
		"updated_at": "` + updatedAt + `",
		"renews_at": "2026-02-01T00:00:00Z"
	}`
}

func subscriptionEvent(attributes string) []byte {
	return []byte(`{
		"meta": {"custom_data": {"app_id": "app-123", "user_id": "user-456"}},
		"data": {"id": "ls-sub-789", "attributes": ` + attributes + `}
	}`)
}

func TestBillingWebhookRejectsEmptySigningSecret(t *testing.T) {
	server := newTestSubscriptionWebhookServer(t, "")

	payload := subscriptionEvent(subscriptionAttributes("1", "2026-01-01T00:00:00Z"))

	// An attacker computes HMAC using an empty key, which would succeed if not failed-closed
	attackerSignature := computeSignature("", payload)

	rec := postWebhookRequest(server, "subscription_created", attackerSignature, payload)
	assertWebhookUnauthorized(t, rec)
}

func TestBillingWebhookRejectsInvalidSignature(t *testing.T) {
	server := newTestSubscriptionWebhookServer(t, testSecret)

	payload := subscriptionEvent(subscriptionAttributes("1", "2026-01-01T00:00:00Z"))

	rec := postWebhookRequest(server, "subscription_created", "invalid-or-tampered-signature", payload)
	assertWebhookUnauthorized(t, rec)
}

func assertWebhookUnauthorized(t *testing.T, rec *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp wire.APIResponse[struct{}]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "invalid_signature", resp.Error.Code)
}

func TestBillingWebhookAcknowledgesUnrelatedEvents(t *testing.T) {
	server := newTestSubscriptionWebhookServer(t, testSecret)

	payload := []byte(`{"data":{"type":"orders","id":"order-1"}}`)
	rec := postWebhookRequest(server, "order_created", computeSignature(testSecret, payload), payload)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp wire.APIResponse[wire.BillingWebhookResponse]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestBillingWebhookSyncsCurrentSubscription(t *testing.T) {
	server := newTestSubscriptionWebhookServer(t, testSecret)
	server.lsAttributes = subscriptionAttributes("1", "2026-01-01T00:00:10Z")

	// Right after checkout LemonSqueezy sends an event without a product, the
	// current state has to be fetched instead.
	payload := subscriptionEvent(subscriptionAttributes("0", "2026-01-01T00:00:00Z"))
	rec := postWebhookRequest(server, "subscription_updated", computeSignature(testSecret, payload), payload)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	require.NotNil(t, server.subStore.upserted)
	assert.Equal(t, "user-456", server.subStore.upserted.UserID)
	assert.Equal(t, "ls-sub-789", server.subStore.upserted.LemonsqueezySubscriptionID.String)
	assert.Equal(t, "1", server.subStore.upserted.LemonsqueezyProductID.String)

	require.NotNil(t, server.entStore.upserted)
	assert.Equal(t, "app-123", server.entStore.upserted.AppID)
	assert.Equal(t, "pro", server.entStore.upserted.PlanID)
}

func TestBillingWebhookKeepsNewerSubscription(t *testing.T) {
	server := newTestSubscriptionWebhookServer(t, testSecret)
	newer := time.Date(2026, 1, 1, 0, 0, 20, 0, time.UTC)
	renewsAt := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	server.subStore.upserted = &model.Subscription{
		ID:                         "sub-1",
		UserID:                     "user-456",
		UpdatedAt:                  newer,
		RenewsAt:                   null.TimeFrom(renewsAt),
		LemonsqueezySubscriptionID: null.StringFrom("ls-sub-789"),
		LemonsqueezyProductID:      null.StringFrom("1"),
	}

	// A concurrent sync stored a newer state after this one was fetched
	server.lsAttributes = subscriptionAttributes("1", "2026-01-01T00:00:10Z")
	payload := subscriptionEvent(server.lsAttributes)
	rec := postWebhookRequest(server, "subscription_updated", computeSignature(testSecret, payload), payload)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	assert.Equal(t, newer, server.subStore.upserted.UpdatedAt)
	require.NotNil(t, server.entStore.upserted)
	assert.Equal(t, "app-123", server.entStore.upserted.AppID)
	assert.Equal(t, null.TimeFrom(renewsAt), server.entStore.upserted.EndsAt)
}

func TestBillingWebhookKeepsProductOnIncompleteSubscription(t *testing.T) {
	server := newTestSubscriptionWebhookServer(t, testSecret)
	server.subStore.upserted = &model.Subscription{
		UserID:                     "user-456",
		LemonsqueezySubscriptionID: null.StringFrom("ls-sub-789"),
		LemonsqueezyProductID:      null.StringFrom("1"),
	}

	server.lsAttributes = subscriptionAttributes("0", "2026-01-01T00:00:10Z")
	payload := subscriptionEvent(server.lsAttributes)
	postWebhookRequest(server, "subscription_updated", computeSignature(testSecret, payload), payload)

	assert.Equal(t, "1", server.subStore.upserted.LemonsqueezyProductID.String)
}
