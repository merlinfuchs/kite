package billing

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func lemonSqueezySubscription(renewsAt time.Time) lemonsqueezy.Subscription {
	return lemonsqueezy.Subscription{
		ProductName: "Pro",
		Status:      "active",
		RenewsAt:    renewsAt,
	}
}

func TestEntitlementEndsAt(t *testing.T) {
	renews := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	ends := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	t.Run("renews_at is used while the subscription keeps renewing", func(t *testing.T) {
		got := entitlementEndsAt(null.TimeFrom(renews), null.Time{})
		assert.Equal(t, null.TimeFrom(renews), got)
	})

	t.Run("ends_at wins when the subscription ends before the next renewal", func(t *testing.T) {
		got := entitlementEndsAt(null.TimeFrom(renews), null.TimeFrom(ends))
		assert.Equal(t, null.TimeFrom(ends), got)
	})

	t.Run("renews_at wins when it comes first", func(t *testing.T) {
		later := ends.AddDate(0, 1, 0)
		got := entitlementEndsAt(null.TimeFrom(ends), null.TimeFrom(later))
		assert.Equal(t, null.TimeFrom(ends), got)
	})

	t.Run("ends_at is used when the subscription will not renew", func(t *testing.T) {
		got := entitlementEndsAt(null.Time{}, null.TimeFrom(ends))
		assert.Equal(t, null.TimeFrom(ends), got)
	})

	t.Run("a subscription with neither date ends the entitlement now", func(t *testing.T) {
		before := time.Now().UTC()
		got := entitlementEndsAt(null.Time{}, null.Time{})

		// A paused subscription reports neither date. The entitlement has to
		// land in the past so ActiveEntitlements stops returning it.
		assert.True(t, got.Valid)
		assert.False(t, got.Time.Before(before))
		assert.False(t, got.Time.After(time.Now().UTC()))
	})
}

func TestSubscriptionFromLemonSqueezy(t *testing.T) {
	t.Run("a zero renews_at is treated as absent", func(t *testing.T) {
		// The API client decodes a null renews_at into the zero time, which must
		// not reach the database as a real date.
		got := SubscriptionFromLemonSqueezy("123", lemonSqueezySubscription(time.Time{}), "user-1")
		assert.False(t, got.RenewsAt.Valid)
	})

	t.Run("a real renews_at is kept", func(t *testing.T) {
		renews := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		got := SubscriptionFromLemonSqueezy("123", lemonSqueezySubscription(renews), "user-1")
		assert.Equal(t, null.TimeFrom(renews), got.RenewsAt)
	})
}

// paddleSubscriptionJSON is a subscription like the Paddle API returns it, to
// Premium, with extra fields like `"next_billed_at": "..."` added.
func paddleSubscriptionJSON(id string, status string, updatedAt time.Time, extra string) string {
	if extra != "" {
		extra = "," + extra
	}
	return `{
		"id": "` + id + `",
		"status": "` + status + `",
		"customer_id": "ctm_1",
		"created_at": "2026-11-01T10:00:00.123Z",
		"updated_at": "` + updatedAt.Format(time.RFC3339Nano) + `",
		"items": [{
			"status": "active",
			"quantity": 1,
			"price": {"id": "pri_yearly", "product_id": "pro_1"},
			"product": {"id": "pro_1", "name": "Premium"}
		}],
		"custom_data": {"user_id": "user-1", "app_id": "app-1"}` + extra + `
	}`
}

func TestSubscriptionFromPaddle(t *testing.T) {
	updated := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	next := time.Date(2026, 12, 1, 10, 0, 0, 0, time.UTC)

	from := func(t *testing.T, status string, extra string) model.Subscription {
		var sub paddle.Subscription
		require.NoError(t, json.Unmarshal([]byte(paddleSubscriptionJSON("sub_1", status, updated, extra)), &sub))
		got, err := SubscriptionFromPaddle(&sub, "user-1")
		require.NoError(t, err)
		return got
	}

	t.Run("an active subscription renews", func(t *testing.T) {
		got := from(t, "active", `"next_billed_at": "2026-12-01T10:00:00Z"`)
		assert.Equal(t, model.Subscription{
			UserID:               "user-1",
			Source:               model.SubscriptionSourcePaddle,
			DisplayName:          "Premium",
			Status:               model.SubscriptionStatusActive,
			StatusFormatted:      "Active",
			RenewsAt:             null.TimeFrom(next),
			CreatedAt:            time.Date(2026, 11, 1, 10, 0, 0, 123000000, time.UTC),
			UpdatedAt:            updated,
			PaddleSubscriptionID: null.StringFrom("sub_1"),
			PaddleCustomerID:     null.StringFrom("ctm_1"),
			PaddleProductID:      null.StringFrom("pro_1"),
			PaddlePriceID:        null.StringFrom("pri_yearly"),
		}, got)
		assert.True(t, got.IsActive())
	})

	t.Run("a cancellation at the end of the period ends it then", func(t *testing.T) {
		got := from(t, "active", `"scheduled_change": {"action": "cancel", "effective_at": "2026-12-01T10:00:00Z"}`)
		assert.Equal(t, model.SubscriptionStatusCancelled, got.Status)
		assert.Equal(t, null.TimeFrom(next), got.EndsAt)
		assert.Equal(t, null.TimeFrom(next), got.PaidUntil())
	})

	t.Run("a pause at the end of the period ends it then", func(t *testing.T) {
		got := from(t, "active", `"scheduled_change": {"action": "pause", "effective_at": "2026-12-01T10:00:00Z"}`)
		assert.Equal(t, model.SubscriptionStatusActive, got.Status)
		assert.Equal(t, null.TimeFrom(next), got.EndsAt)
	})

	t.Run("a scheduled resume doesn't end it", func(t *testing.T) {
		got := from(t, "paused", `"scheduled_change": {"action": "resume", "effective_at": "2026-12-01T10:00:00Z"}`)
		assert.Equal(t, model.SubscriptionStatusPaused, got.Status)
		assert.False(t, got.PaidUntil().Valid)
	})

	t.Run("a canceled subscription has expired", func(t *testing.T) {
		got := from(t, "canceled", `"canceled_at": "2026-11-02T09:00:00Z"`)
		assert.Equal(t, model.SubscriptionStatusExpired, got.Status)
		assert.Equal(t, null.TimeFrom(time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)), got.EndsAt)
		assert.False(t, got.IsActive())
	})

	t.Run("a past due subscription still gives access", func(t *testing.T) {
		got := from(t, "past_due", `"next_billed_at": "2026-12-01T10:00:00Z"`)
		assert.Equal(t, model.SubscriptionStatusPastDue, got.Status)
		assert.True(t, got.IsActive())
	})

	t.Run("a trial ends with its billing period", func(t *testing.T) {
		got := from(t, "trialing", `"current_billing_period": {"starts_at": "2026-11-01T10:00:00Z", "ends_at": "2026-12-01T10:00:00Z"}`)
		assert.Equal(t, model.SubscriptionStatusOnTrial, got.Status)
		assert.Equal(t, null.TimeFrom(next), got.TrialEndsAt)
	})

	t.Run("an unknown status doesn't give access", func(t *testing.T) {
		got := from(t, "frozen", "")
		assert.Equal(t, "frozen", got.Status)
		assert.False(t, got.IsActive())
	})

	t.Run("a subscription without items fails", func(t *testing.T) {
		_, err := SubscriptionFromPaddle(&paddle.Subscription{ID: "sub_1"}, "user-1")
		assert.Error(t, err)
	})
}
