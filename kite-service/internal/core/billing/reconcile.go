package billing

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// ReconcileDelay is the pause between API calls. It keeps reconciliation well
// inside the rate limits of 300 requests per minute of LemonSqueezy and 240 of
// Paddle, which webhooks and checkouts share.
const ReconcileDelay = time.Second

// Run reconciles on startup and every few hours after, which recovers
// webhooks that failed for longer than they're retried.
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
	// Skipped counts subscriptions we don't look up, because their provider
	// isn't configured or they have expired, which is final.
	Skipped int
	// Changed counts subscriptions that were updated by their provider since
	// we last stored them.
	Changed int
	// Unchanged counts subscriptions that were already up to date.
	Unchanged int
	// Failed counts subscriptions we could not reconcile.
	Failed int
}

// Reconcile re-syncs every live subscription from its provider. Unchanged
// ones are synced too, which repairs entitlements a failed sync left behind.
//
// It exists because LemonSqueezy has no bulk webhook replay: events missed
// while the webhook was rejecting them are gone, so the only way to recover is
// to ask for the current state of each subscription. Paddle can replay them,
// but checking costs little.
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
		if !m.canReconcile(subscription) || subscription.Status == model.SubscriptionStatusExpired {
			result.Skipped++
			continue
		}

		// The providers rate limit their APIs, so pace the requests
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(delay):
		}

		log := slog.With(
			slog.String("subscription_id", subscription.ID),
			slog.String("source", subscription.Source.String()),
			slog.String("provider_subscription_id", subscription.ProviderID()),
		)

		updated, err := m.fetch(ctx, subscription)
		if err != nil {
			log.Error("Failed to get subscription from its provider", slog.String("error", err.Error()))
			result.Failed++
			continue
		}

		changed := !updated.UpdatedAt.Equal(subscription.UpdatedAt)

		// LemonSqueezy also bumps updated_at without a change worth logging,
		// around payments for example
		if updated.Status != subscription.Status {
			log.Info(
				"Subscription status is out of date",
				slog.String("stored_status", subscription.Status),
				slog.String("provider_status", updated.Status),
			)
		}

		if !dryRun {
			// The app ID is left empty on purpose, see the doc comment.
			if _, err := m.Sync(ctx, updated, ""); err != nil {
				log.Error("Failed to reconcile subscription", slog.String("error", err.Error()))
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

func (m *SubscriptionManager) canReconcile(sub *model.Subscription) bool {
	if sub.ProviderID() == "" {
		return false
	}
	if sub.Source == model.SubscriptionSourcePaddle {
		return m.paddle != nil
	}
	return m.client != nil
}

// fetch gets the subscription as its provider reports it now.
func (m *SubscriptionManager) fetch(ctx context.Context, sub *model.Subscription) (model.Subscription, error) {
	switch sub.Source {
	case model.SubscriptionSourceLemonSqueezy:
		res, _, err := m.client.Subscriptions.Get(ctx, sub.LemonsqueezySubscriptionID.String)
		if err != nil {
			return model.Subscription{}, err
		}
		return SubscriptionFromLemonSqueezy(res.Data.ID, res.Data.Attributes, sub.UserID), nil
	case model.SubscriptionSourcePaddle:
		res, err := m.getPaddleSubscription(ctx, sub.PaddleSubscriptionID.String)
		if err != nil {
			return model.Subscription{}, err
		}
		return SubscriptionFromPaddle(res, sub.UserID)
	}
	return model.Subscription{}, fmt.Errorf("unknown subscription source %q", sub.Source)
}
