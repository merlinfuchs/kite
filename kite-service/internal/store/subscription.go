package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type SubscriptionStore interface {
	Subscriptions(ctx context.Context, userID string) ([]*model.Subscription, error)
	SubscriptionsByAppID(ctx context.Context, appID string) ([]*model.Subscription, error)
	AllSubscriptions(ctx context.Context) ([]*model.Subscription, error)
	Subscription(ctx context.Context, subscriptionID string) (*model.Subscription, error)
	UpsertLemonSqueezySubscription(ctx context.Context, sub model.Subscription) (*model.Subscription, error)
}
