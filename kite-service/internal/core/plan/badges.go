package plan

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// earlySupporterCutoff is when Kite Plus and the new Premium features
// launched. Users who paid for a plan before are early supporters.
var earlySupporterCutoff = time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC)

// earlyCreatorCutoff is when the badges launched. Users who listed a template
// in the marketplace before, without knowing about the badges, are early
// creators.
var earlyCreatorCutoff = time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)

// creatorMilestones are how many other users must have added any of a
// creator's templates for each creator badge, highest first. The lowest is
// what the creator role in the Kite server needs too.
var creatorMilestones = []int{1000, 100, creatorMinInstallers}

// UserBadges returns the badges of the users by their ID. Users without a
// badge aren't in it.
func (m *PlanManager) UserBadges(ctx context.Context, userIDs []string) (map[string][]model.UserBadge, error) {
	if len(userIDs) == 0 {
		return map[string][]model.UserBadge{}, nil
	}

	grants, err := m.userStore.UserBadgeGrantsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user badge grants: %w", err)
	}
	subs, err := m.subscriptionStore.SubscriptionsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	stats, err := m.bundleStore.CreatorStats(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator stats: %w", err)
	}

	var admins []string
	for _, id := range m.config.AdminUserIDs {
		if slices.Contains(userIDs, id) {
			admins = append(admins, id)
		}
	}

	return userBadges(admins, grants, subs, stats, func(productID string) string {
		if plan := m.PlanByLemonSqueezyProductID(productID); plan != nil {
			return plan.Title
		}
		return ""
	}), nil
}

// userBadges works out the badges of users from whether they're admins, the
// badges admins gave them, their subscriptions and their templates in the
// marketplace. planTitle is the title of the plan of a product, or empty if it
// isn't one.
func userBadges(
	admins []string,
	grants []*model.UserBadgeGrant,
	subs []*model.Subscription,
	stats []*model.CreatorStats,
	planTitle func(productID string) string,
) map[string][]model.UserBadge {
	granted := map[string]map[model.UserBadgeKind]bool{}
	grant := func(userID string, kind model.UserBadgeKind) {
		if granted[userID] == nil {
			granted[userID] = map[model.UserBadgeKind]bool{}
		}
		granted[userID][kind] = true
	}
	// Admins are staff without a grant, and once if they have one too.
	for _, userID := range admins {
		grant(userID, model.UserBadgeStaff)
	}
	for _, g := range grants {
		grant(g.UserID, g.Kind)
	}

	early := map[string]bool{}
	// The plan of the active subscription of each user, the newest if they
	// have several, as subs is newest first.
	plan := map[string]string{}
	for _, sub := range subs {
		title := planTitle(sub.LemonsqueezyProductID.String)
		if title == "" {
			continue
		}
		if sub.CreatedAt.Before(earlySupporterCutoff) {
			early[sub.UserID] = true
		}
		if sub.IsActive() && plan[sub.UserID] == "" {
			plan[sub.UserID] = title
		}
	}

	res := map[string][]model.UserBadge{}
	add := func(userID string, b model.UserBadge) {
		res[userID] = append(res[userID], b)
	}

	// Badges are in the same order for everyone: staff, what admins gave,
	// then support, then templates.
	for userID, kinds := range granted {
		for _, kind := range model.GrantableUserBadges {
			if kinds[kind] {
				add(userID, model.UserBadge{Kind: kind})
			}
		}
	}
	for userID := range early {
		add(userID, model.UserBadge{Kind: model.UserBadgeEarlySupporter})
	}
	for userID, title := range plan {
		add(userID, model.UserBadge{Kind: model.UserBadgeSubscriber, Plan: title})
	}
	for _, s := range stats {
		if s.FirstListedAt.Before(earlyCreatorCutoff) {
			add(s.UserID, model.UserBadge{Kind: model.UserBadgeEarlyCreator})
		}
		// Listing a template is the first level, without a milestone.
		badge := model.UserBadge{Kind: model.UserBadgeCreator}
		for _, milestone := range creatorMilestones {
			if s.InstallerCount >= milestone {
				badge.Milestone = milestone
				break
			}
		}
		add(s.UserID, badge)
	}
	return res
}
