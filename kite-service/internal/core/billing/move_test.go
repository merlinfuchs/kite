package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type moveSubscriptionStore struct {
	store.SubscriptionStore
	lemonSqueezy *model.Subscription
	paddle       *model.Subscription
}

func (s *moveSubscriptionStore) SubscriptionByLemonSqueezyID(ctx context.Context, id string) (*model.Subscription, error) {
	if s.lemonSqueezy == nil || s.lemonSqueezy.LemonsqueezySubscriptionID.String != id {
		return nil, store.ErrNotFound
	}
	return s.lemonSqueezy, nil
}

func (s *moveSubscriptionStore) UpsertLemonSqueezySubscription(ctx context.Context, sub model.Subscription) (*model.Subscription, error) {
	s.lemonSqueezy = &sub
	return &sub, nil
}

func (s *moveSubscriptionStore) UpsertPaddleSubscription(ctx context.Context, sub model.Subscription) (*model.Subscription, error) {
	s.paddle = &sub
	return &sub, nil
}

type moveEntitlementStore struct {
	store.EntitlementStore
}

func (moveEntitlementStore) UpsertSubscriptionEntitlement(ctx context.Context, e model.Entitlement) (*model.Entitlement, error) {
	return &e, nil
}

// moveAPI stubs the APIs of Paddle and Lemon Squeezy for a move of the Lemon
// Squeezy subscription ls_1 to the Paddle subscription sub_1.
type moveAPI struct {
	requests []string
	// paddle is the Paddle subscription, which PATCH requests change.
	paddle map[string]any
	// lemonSqueezyStatus is the status of the Lemon Squeezy subscription.
	lemonSqueezyStatus string
}

var moveRenewsAt = time.Date(2027, 1, 15, 12, 0, 0, 0, time.UTC)

func newMoveAPI(t *testing.T) (*moveAPI, *httptest.Server) {
	a := &moveAPI{
		paddle: map[string]any{
			"id":             "sub_1",
			"status":         "active",
			"customer_id":    "ctm_1",
			"created_at":     "2026-11-28T10:00:00Z",
			"updated_at":     "2026-11-28T10:00:00Z",
			"next_billed_at": "2026-12-28T10:00:00Z",
			"items":          moveItems("pri_zero"),
			"custom_data": map[string]any{
				"user_id":                    "user-1",
				"app_id":                     "app-1",
				"moved_from_lemonsqueezy_id": "ls_1",
				"move_price_id":              "pri_plus",
			},
		},
		lemonSqueezyStatus: "active",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		a.requests = append(a.requests, r.Method+" "+r.URL.Path+" "+string(bytes.TrimSpace(body)))

		switch r.Method + " " + r.URL.Path {
		case "GET /subscriptions/sub_1":
		case "PATCH /subscriptions/sub_1":
			var patch map[string]any
			require.NoError(t, json.Unmarshal(body, &patch))
			if items, ok := patch["items"].([]any); ok {
				a.paddle["items"] = moveItems(items[0].(map[string]any)["price_id"].(string))
			}
			for _, key := range []string{"next_billed_at", "custom_data"} {
				if v, ok := patch[key]; ok {
					a.paddle[key] = v
				}
			}
		case "DELETE /v1/subscriptions/ls_1":
			a.lemonSqueezyStatus = "cancelled"
			fallthrough
		case "GET /v1/subscriptions/ls_1":
			attrs := map[string]any{
				"status":     a.lemonSqueezyStatus,
				"product_id": 1,
				"variant_id": 10,
				"renews_at":  moveRenewsAt,
				"created_at": "2026-10-15T12:00:00Z",
				"updated_at": "2026-11-28T10:00:00Z",
			}
			if a.lemonSqueezyStatus == "cancelled" {
				attrs["ends_at"] = moveRenewsAt
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": "ls_1", "attributes": attrs}})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": a.paddle})
	}))
	t.Cleanup(server.Close)
	return a, server
}

func moveItems(priceID string) []any {
	return []any{map[string]any{
		"status":   "active",
		"quantity": 1,
		"price":    map[string]any{"id": priceID, "product_id": "pro_plus"},
		"product":  map[string]any{"id": "pro_plus", "name": "Plus"},
	}}
}

func newMoveManager(t *testing.T, owner string) (*SubscriptionManager, *moveSubscriptionStore, *moveAPI) {
	api, server := newMoveAPI(t)

	subStore := &moveSubscriptionStore{lemonSqueezy: &model.Subscription{
		UserID:                     owner,
		Source:                     model.SubscriptionSourceLemonSqueezy,
		Status:                     model.SubscriptionStatusActive,
		RenewsAt:                   null.TimeFrom(moveRenewsAt),
		LemonsqueezySubscriptionID: null.StringFrom("ls_1"),
	}}
	plans := []model.Plan{{
		ID:                    "plus",
		LemonSqueezyProductID: "1",
		LemonSqueezyVariantID: "10",
		PaddleProductID:       "pro_plus",
		PaddlePriceID:         "pri_plus",
	}}
	planManager := plan.NewPlanManager(nil, subStore, nil, nil, plans, plan.PlanManagerConfig{})

	paddleClient, err := paddle.New("key", paddle.WithBaseURL(server.URL))
	require.NoError(t, err)
	lsClient := lemonsqueezy.New(lemonsqueezy.WithBaseURL(server.URL), lemonsqueezy.WithAPIKey("key"))

	return NewSubscriptionManager(subStore, moveEntitlementStore{}, planManager, lsClient, paddleClient), subStore, api
}

func TestMoveFromLemonSqueezy(t *testing.T) {
	m, subStore, api := newMoveManager(t, "user-1")

	sub, err := m.RefreshPaddle(context.Background(), "sub_1")
	require.NoError(t, err)

	require.Len(t, api.requests, 5)
	assert.Equal(t, "GET /subscriptions/sub_1 ", api.requests[0])
	assert.Equal(t, `PATCH /subscriptions/sub_1 {"items":[{"price_id":"pri_plus","quantity":1}],"proration_billing_mode":"do_not_bill"}`, api.requests[1])
	assert.Equal(t, "GET /v1/subscriptions/ls_1 ", api.requests[2])
	assert.Equal(t, "DELETE /v1/subscriptions/ls_1 ", api.requests[3])
	assert.Equal(t, `PATCH /subscriptions/sub_1 {"next_billed_at":"2027-01-15T12:00:00Z","custom_data":{"app_id":"app-1","move_price_id":"pri_plus","moved":true,"moved_from_lemonsqueezy_id":"ls_1","user_id":"user-1"},"proration_billing_mode":"do_not_bill"}`, api.requests[4])

	// The first payment on Paddle is when Lemon Squeezy would have renewed,
	// which serves out its paid period until then.
	assert.Equal(t, "pri_plus", sub.PaddlePriceID.String)
	assert.Equal(t, null.TimeFrom(moveRenewsAt), sub.RenewsAt)
	assert.Equal(t, model.SubscriptionStatusCancelled, subStore.lemonSqueezy.Status)
	assert.Equal(t, null.TimeFrom(moveRenewsAt), subStore.lemonSqueezy.EndsAt)

	t.Run("is done once", func(t *testing.T) {
		api.requests = nil
		_, err := m.RefreshPaddle(context.Background(), "sub_1")
		require.NoError(t, err)
		assert.Equal(t, []string{"GET /subscriptions/sub_1 "}, api.requests)
	})
}

func TestMoveFromLemonSqueezyRetry(t *testing.T) {
	// A webhook before failed after cancelling on Lemon Squeezy.
	m, _, api := newMoveManager(t, "user-1")
	api.paddle["items"] = moveItems("pri_plus")
	api.lemonSqueezyStatus = "cancelled"

	_, err := m.RefreshPaddle(context.Background(), "sub_1")
	require.NoError(t, err)
	require.Len(t, api.requests, 3)
	assert.Equal(t, "GET /v1/subscriptions/ls_1 ", api.requests[1])
	assert.Contains(t, api.requests[2], `"next_billed_at":"2027-01-15T12:00:00Z"`)
}

func TestMoveFromLemonSqueezyOfAnotherUser(t *testing.T) {
	m, _, api := newMoveManager(t, "user-2")

	_, err := m.RefreshPaddle(context.Background(), "sub_1")
	assert.ErrorIs(t, err, ErrInvalidMove)
	assert.Len(t, api.requests, 1)
}
