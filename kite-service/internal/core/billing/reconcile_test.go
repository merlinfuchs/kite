package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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

type fakeSubscriptionStore struct {
	store.SubscriptionStore
	subscriptions []*model.Subscription
	upserted      []string
}

func (s *fakeSubscriptionStore) AllSubscriptions(ctx context.Context) ([]*model.Subscription, error) {
	return s.subscriptions, nil
}

func (s *fakeSubscriptionStore) UpsertLemonSqueezySubscription(ctx context.Context, sub model.Subscription) (*model.Subscription, error) {
	s.upserted = append(s.upserted, sub.LemonsqueezySubscriptionID.String)
	return &sub, nil
}

func (s *fakeSubscriptionStore) UpsertPaddleSubscription(ctx context.Context, sub model.Subscription) (*model.Subscription, error) {
	s.upserted = append(s.upserted, sub.PaddleSubscriptionID.String)
	return &sub, nil
}

type fakeEntitlementStore struct {
	store.EntitlementStore
}

func (s *fakeEntitlementStore) UpdateSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) error {
	return nil
}

func TestReconcile(t *testing.T) {
	stored := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var mu sync.Mutex
	var requested []string
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/subscriptions/")
		mu.Lock()
		requested = append(requested, id)
		mu.Unlock()

		updatedAt := stored
		if id == "changed" {
			updatedAt = stored.Add(time.Minute)
		}
		_, _ = w.Write([]byte(`{"data":{"id":"` + id + `","attributes":{"product_id":1,"status":"active","updated_at":"` + updatedAt.Format(time.RFC3339) + `"}}}`))
	}))
	defer api.Close()

	paddleAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/subscriptions/")
		mu.Lock()
		requested = append(requested, id)
		mu.Unlock()

		_, _ = w.Write([]byte(`{"data":` + paddleSubscriptionJSON(id, "canceled", stored.Add(time.Minute), "") + `}`))
	}))
	defer paddleAPI.Close()

	subscription := func(id string, status string) *model.Subscription {
		return &model.Subscription{
			ID:                         id,
			Source:                     model.SubscriptionSourceLemonSqueezy,
			Status:                     status,
			UpdatedAt:                  stored,
			LemonsqueezySubscriptionID: null.StringFrom(id),
		}
	}
	subStore := &fakeSubscriptionStore{subscriptions: []*model.Subscription{
		subscription("changed", "active"),
		subscription("unchanged", "active"),
		subscription("expired", "expired"),
		{ID: "manual", Status: "active"},
		{
			ID:                   "paddle",
			Source:               model.SubscriptionSourcePaddle,
			Status:               "active",
			UpdatedAt:            stored,
			PaddleSubscriptionID: null.StringFrom("sub_1"),
		},
	}}
	entStore := &fakeEntitlementStore{}

	plans := []model.Plan{{ID: "pro", LemonSqueezyProductID: "1", PaddleProductID: "pro_1"}}
	planManager := plan.NewPlanManager(entStore, subStore, nil, nil, plans, plan.PlanManagerConfig{})
	client := lemonsqueezy.New(lemonsqueezy.WithBaseURL(api.URL), lemonsqueezy.WithAPIKey("key"))
	paddleClient, err := paddle.New("key", paddle.WithBaseURL(paddleAPI.URL))
	require.NoError(t, err)
	m := NewSubscriptionManager(subStore, entStore, planManager, client, paddleClient)

	res, err := m.Reconcile(context.Background(), false, 0)
	require.NoError(t, err)

	assert.Equal(t, ReconcileResult{Total: 5, Skipped: 2, Changed: 2, Unchanged: 1}, res)
	mu.Lock()
	assert.Equal(t, []string{"changed", "unchanged", "sub_1"}, requested)
	mu.Unlock()
	// Unchanged ones are synced too, to repair their entitlements
	assert.Equal(t, []string{"changed", "unchanged", "sub_1"}, subStore.upserted)

	t.Run("Lemon Squeezy subscriptions are skipped without its key", func(t *testing.T) {
		subStore.upserted = nil
		m := NewSubscriptionManager(subStore, entStore, planManager, nil, paddleClient)

		res, err := m.Reconcile(context.Background(), false, 0)
		require.NoError(t, err)
		assert.Equal(t, 4, res.Skipped)
		assert.Equal(t, []string{"sub_1"}, subStore.upserted)
	})

	t.Run("Paddle subscriptions are skipped without Paddle", func(t *testing.T) {
		subStore.upserted = nil
		m := NewSubscriptionManager(subStore, entStore, planManager, client, nil)

		res, err := m.Reconcile(context.Background(), false, 0)
		require.NoError(t, err)
		assert.Equal(t, 3, res.Skipped)
		assert.Equal(t, []string{"changed", "unchanged"}, subStore.upserted)
	})
}
