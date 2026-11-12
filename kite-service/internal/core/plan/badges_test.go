package plan

import (
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func TestUserBadges(t *testing.T) {
	before := earlySupporterCutoff.Add(-24 * time.Hour)
	after := earlyCreatorCutoff.Add(24 * time.Hour)
	sub := func(userID, product, status string, createdAt time.Time) *model.Subscription {
		return &model.Subscription{UserID: userID, LemonsqueezyProductID: null.StringFrom(product), Status: status, CreatedAt: createdAt}
	}
	titles := map[string]string{"premium": "Premium", "plus": "Plus"}

	got := userBadges(
		[]*model.UserBadgeGrant{{UserID: "dev", Kind: model.UserBadgeContributor}},
		[]*model.Subscription{
			// Newest first, like the store returns them.
			sub("fan", "plus", model.SubscriptionStatusActive, after),
			sub("fan", "premium", "expired", before),
			sub("gone", "premium", "expired", before),
			sub("new", "premium", model.SubscriptionStatusActive, after),
			sub("donor", "donation", model.SubscriptionStatusActive, before),
		},
		[]*model.CreatorStats{
			{UserID: "dev", FirstListedAt: before, InstallerCount: 100},
			{UserID: "late", FirstListedAt: after, InstallerCount: 9},
			{UserID: "big", FirstListedAt: after, InstallerCount: 1000},
		},
		func(productID string) string { return titles[productID] },
	)

	want := map[string][]model.UserBadge{
		"dev": {
			{Kind: model.UserBadgeContributor},
			{Kind: model.UserBadgeEarlyCreator},
			{Kind: model.UserBadgeCreator, Milestone: 100},
		},
		"fan": {
			{Kind: model.UserBadgeEarlySupporter},
			{Kind: model.UserBadgeSubscriber, Plan: "Plus"},
		},
		"gone": {{Kind: model.UserBadgeEarlySupporter}},
		"new":  {{Kind: model.UserBadgeSubscriber, Plan: "Premium"}},
		"big":  {{Kind: model.UserBadgeCreator, Milestone: 1000}},
	}
	if len(got) != len(want) {
		t.Fatalf("got badges for %d users, want %d: %v", len(got), len(want), got)
	}
	for userID, badges := range want {
		if len(got[userID]) != len(badges) {
			t.Fatalf("%s: got %v, want %v", userID, got[userID], badges)
		}
		for i := range badges {
			if got[userID][i] != badges[i] {
				t.Errorf("%s: got %v, want %v", userID, got[userID], badges)
			}
		}
	}
}
