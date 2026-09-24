package billing

import (
	"context"
	"log/slog"
	"time"
)

// ReconcileResult summarises a reconciliation run.
type ReconcileResult struct {
	// Total is the number of subscriptions considered.
	Total int
	// Skipped counts subscriptions we cannot look up, because they were never
	// bought through LemonSqueezy.
	Skipped int
	// Changed counts subscriptions whose stored status no longer matched
	// LemonSqueezy.
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
		if !subscription.LemonsqueezySubscriptionID.Valid {
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

		changed := res.Data.Attributes.Status != subscription.Status
		if changed {
			slog.Info(
				"Subscription status is out of date",
				slog.String("subscription_id", subscription.ID),
				slog.String("ls_subscription_id", lsID),
				slog.String("stored_status", subscription.Status),
				slog.String("lemonsqueezy_status", res.Data.Attributes.Status),
			)
		}

		if !dryRun {
			updated := SubscriptionFromLemonSqueezy(res.Data.ID, res.Data.Attributes)
			updated.UserID = subscription.UserID

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
