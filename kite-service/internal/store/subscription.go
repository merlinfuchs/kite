package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type SubscriptionStore interface {
	Subscriptions(ctx context.Context, userID string) ([]*model.Subscription, error)
	UpsertLemonSqueezySubscription(ctx context.Context, sub model.Subscription) (*model.Subscription, error)
}
