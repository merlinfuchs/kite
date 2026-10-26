package billing

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// ReconcileDelay is the pause between LemonSqueezy API calls. It keeps
// reconciliation well inside the rate limit of 300 requests per minute, which
// webhooks and checkouts share.
const ReconcileDelay = time.Second

// Run reconciles on startup and every few hours after, which recovers
// webhooks that failed for longer than LemonSqueezy retries them.
func (m *SubscriptionManager) Run(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for {
			if _, err := m.Reconcile(ctx, false, ReconcileDelay); err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("Failed to reconcile subscriptions", slog.String("error", err.Error()))
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// ReconcileResult summarises a reconciliation run.
type ReconcileResult struct {
	// Total is the number of subscriptions considered.
	Total int
	// Skipped counts subscriptions we don't look up, because they were never
	// bought through LemonSqueezy or have expired, which is final.
	Skipped int
	// Changed counts subscriptions that were updated in LemonSqueezy since we
	// last stored them.
	Changed int
	// Unchanged counts subscriptions that were already up to date.
	Unchanged int
	// Failed counts subscriptions we could not reconcile.
	Failed int
}

// Reconcile re-syncs every live subscription from LemonSqueezy. Unchanged
// ones are synced too, which repairs entitlements a failed sync left behind.
//
// It exists because LemonSqueezy has no bulk webhook replay: events missed
// while the webhook was rejecting them are gone, so the only way to recover is
// to ask for the current state of each subscription.
//
// Entitlements are updated through the no-app-ID path of Sync, which updates
// every entitlement a subscription already holds. That deliberately does not
// create missing entitlements: a subscription whose original webhook never
// landed has no app to attribute one to, and guessing would grant the wrong app
// premium.
func (m *SubscriptionManager) Reconcile(ctx context.Context, dryRun bool, delay time.Duration) (result ReconcileResult, err error) {
	// Also logged on failure, because it says how far the run got
	defer func() {
		slog.Info(
			"Reconciled subscriptions",
			slog.Int("total", result.Total),
			slog.Int("changed", result.Changed),
			slog.Int("unchanged", result.Unchanged),
			slog.Int("skipped", result.Skipped),
			slog.Int("failed", result.Failed),
		)
	}()

	subscriptions, err := m.subscriptionStore.AllSubscriptions(ctx)
	if err != nil {
		return result, err
	}

	result.Total = len(subscriptions)

	for _, subscription := range subscriptions {
		if !subscription.LemonsqueezySubscriptionID.Valid || subscription.Status == model.SubscriptionStatusExpired {
			result.Skipped++
			continue
		}

		// LemonSqueezy rate limits the API, so pace the requests
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(delay):
		}

		lsID := subscription.LemonsqueezySubscriptionID.String

		res, _, err := m.client.Subscriptions.Get(ctx, lsID)
		if err != nil {
			slog.Error(
				"Failed to get subscription from LemonSqueezy",
				slog.String("subscription_id", subscription.ID),
				slog.String("ls_subscription_id", lsID),
				slog.String("error", err.Error()),
			)
			result.Failed++
			continue
		}

		changed := !res.Data.Attributes.UpdatedAt.Equal(subscription.UpdatedAt)

		// LemonSqueezy also bumps updated_at without a change worth logging,
		// around payments for example
		if res.Data.Attributes.Status != subscription.Status {
			slog.Info(
				"Subscription status is out of date",
				slog.String("subscription_id", subscription.ID),
				slog.String("ls_subscription_id", lsID),
				slog.String("stored_status", subscription.Status),
				slog.String("lemonsqueezy_status", res.Data.Attributes.Status),
			)
		}

		if !dryRun {
			updated := SubscriptionFromLemonSqueezy(res.Data.ID, res.Data.Attributes, subscription.UserID)

			// The app ID is left empty on purpose, see the doc comment.
			if _, err := m.Sync(ctx, updated, ""); err != nil {
				slog.Error(
					"Failed to reconcile subscription",
					slog.String("subscription_id", subscription.ID),
					slog.String("ls_subscription_id", lsID),
					slog.String("error", err.Error()),
				)
				result.Failed++
				continue
			}
		}

		if changed {
			result.Changed++
		} else {
			result.Unchanged++
		}
	}

	return result, nil
}
