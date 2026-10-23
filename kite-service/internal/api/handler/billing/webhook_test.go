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

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	corebilling "github.com/kitecloud/kite/kite-service/internal/core/billing"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSubscriptionStore struct {
	store.SubscriptionStore
	upserted *model.Subscription
}

func (s *fakeSubscriptionStore) UpsertLemonSqueezySubscription(ctx context.Context, sub model.Subscription) (*model.Subscription, error) {
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

func TestBillingWebhookRejectsEmptySigningSecret(t *testing.T) {
	h := NewBillingHandler(
		BillingHandlerConfig{LemonSqueezySigningSecret: ""},
		nil, nil, nil, nil,
		corebilling.NewLemonSqueezyClient("key", ""),
	)
	server := newTestWebhookServer(h)

	payload := []byte(`{"meta":{"custom_data":{"app_id":"app-1","user_id":"user-1"}},"data":{"id":"sub-1","attributes":{"product_name":"Pro","product_id":"prod-1"}}}`)

	// An attacker computes HMAC using an empty key, which would succeed if not failed-closed
	attackerSignature := computeSignature("", payload)

	rec := postWebhookRequest(server, "subscription_created", attackerSignature, payload)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp wire.APIResponse[struct{}]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "invalid_signature", resp.Error.Code)
}

func TestBillingWebhookRejectsInvalidSignature(t *testing.T) {
	secret := "valid-secret-key"
	h := NewBillingHandler(
		BillingHandlerConfig{LemonSqueezySigningSecret: secret},
		nil, nil, nil, nil,
		corebilling.NewLemonSqueezyClient("key", secret),
	)
	server := newTestWebhookServer(h)

	payload := []byte(`{"meta":{"custom_data":{"app_id":"app-1","user_id":"user-1"}},"data":{"id":"sub-1","attributes":{"product_name":"Pro","product_id":"prod-1"}}}`)

	rec := postWebhookRequest(server, "subscription_created", "invalid-or-tampered-signature", payload)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp wire.APIResponse[struct{}]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "invalid_signature", resp.Error.Code)
}

func TestBillingWebhookAcknowledgesUnrelatedEvents(t *testing.T) {
	secret := "valid-secret-key"
	h := NewBillingHandler(
		BillingHandlerConfig{LemonSqueezySigningSecret: secret},
		nil, nil, nil, nil,
		corebilling.NewLemonSqueezyClient("key", secret),
	)
	server := newTestWebhookServer(h)

	payload := []byte(`{"data":{"type":"orders","id":"order-1"}}`)
	validSig := computeSignature(secret, payload)

	rec := postWebhookRequest(server, "order_created", validSig, payload)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp wire.APIResponse[wire.BillingWebhookResponse]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestBillingWebhookProcessesValidSubscriptionEvent(t *testing.T) {
	secret := "valid-secret-key"
	subStore := &fakeSubscriptionStore{}
	entStore := &fakeEntitlementStore{}

	plans := []model.Plan{
		{
			ID:                    "pro",
			LemonSqueezyProductID: "1",
		},
	}
	planManager := plan.NewPlanManager(entStore, subStore, nil, plans, plan.PlanManagerConfig{})

	lsClient := corebilling.NewLemonSqueezyClient("key", secret)
	subManager := corebilling.NewSubscriptionManager(subStore, entStore, planManager, lsClient)

	h := NewBillingHandler(
		BillingHandlerConfig{LemonSqueezySigningSecret: secret},
		nil, subStore, planManager, subManager, lsClient,
	)
	server := newTestWebhookServer(h)

	payload := []byte(`{
		"meta": {
			"custom_data": {
				"app_id": "app-123",
				"user_id": "user-456"
			}
		},
		"data": {
			"id": "ls-sub-789",
			"attributes": {
				"product_name": "Pro Plan",
				"product_id": 1,
				"variant_id": 1,
				"order_id": 100,
				"customer_id": 200,
				"status": "active",
				"status_formatted": "Active",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z"
			}
		}
	}`)
	validSig := computeSignature(secret, payload)

	rec := postWebhookRequest(server, "subscription_created", validSig, payload)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	require.NotNil(t, subStore.upserted)
	assert.Equal(t, "user-456", subStore.upserted.UserID)
	assert.Equal(t, "ls-sub-789", subStore.upserted.LemonsqueezySubscriptionID.String)
	assert.Equal(t, "1", subStore.upserted.LemonsqueezyProductID.String)

	require.NotNil(t, entStore.upserted)
	assert.Equal(t, "app-123", entStore.upserted.AppID)
	assert.Equal(t, "pro", entStore.upserted.PlanID)
}
