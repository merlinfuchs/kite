package billing

import (
	"context"
	"log/slog"
	"time"
)

// ReconcileDelay is the pause between LemonSqueezy API calls that keeps
// reconciliation inside the rate limit.
const ReconcileDelay = 250 * time.Millisecond

// Run reconciles every few hours, which recovers webhooks that failed for
// longer than LemonSqueezy retries them.
func (m *SubscriptionManager) Run(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				res, err := m.Reconcile(ctx, false, ReconcileDelay)
				if err != nil {
					slog.Error("Failed to reconcile subscriptions", slog.String("error", err.Error()))
				}
				if res.Changed > 0 || res.Failed > 0 {
					slog.Info(
						"Reconciled subscriptions",
						slog.Int("changed", res.Changed),
						slog.Int("failed", res.Failed),
					)
				}
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

// Reconcile brings every stored subscription back in line with LemonSqueezy.
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
func (m *SubscriptionManager) Reconcile(ctx context.Context, dryRun bool, delay time.Duration) (ReconcileResult, error) {
	var result ReconcileResult

	subscriptions, err := m.subscriptionStore.AllSubscriptions(ctx)
	if err != nil {
		return result, err
	}

	result.Total = len(subscriptions)

	var requests int
	for _, subscription := range subscriptions {
		if !subscription.LemonsqueezySubscriptionID.Valid || subscription.Status == "expired" {
			result.Skipped++
			continue
		}

		// LemonSqueezy rate limits the API, and this runs over every
		// subscription we have, so pace the requests.
		if requests > 0 && delay > 0 {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(delay):
			}
		}

		lsID := subscription.LemonsqueezySubscriptionID.String
		requests++

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

		if res.Data.Attributes.UpdatedAt.Equal(subscription.UpdatedAt) {
			result.Unchanged++
			continue
		}

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

		result.Changed++
	}

	return result, nil
}
