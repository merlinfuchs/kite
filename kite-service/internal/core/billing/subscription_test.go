package billing

import (
	"testing"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/stretchr/testify/assert"
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
		got := SubscriptionFromLemonSqueezy("123", lemonSqueezySubscription(time.Time{}))
		assert.False(t, got.RenewsAt.Valid)
	})

	t.Run("a real renews_at is kept", func(t *testing.T) {
		renews := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		got := SubscriptionFromLemonSqueezy("123", lemonSqueezySubscription(renews))
		assert.Equal(t, null.TimeFrom(renews), got.RenewsAt)
	})
}
