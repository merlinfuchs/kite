package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type EventListenerStore interface {
	EventListenersByApp(ctx context.Context, appID string) ([]*model.EventListener, error)
	CountEventListenersByApp(ctx context.Context, appID string) (int, error)
	CountEventListenersByAppAndSource(ctx context.Context, appID string, source model.EventSource) (int, error)
	UpdateEventListenersLastRunAt(ctx context.Context, lastRunAts map[string]time.Time) error
	EventListener(ctx context.Context, id string) (*model.EventListener, error)
	CreateEventListener(ctx context.Context, eventListener *model.EventListener) (*model.EventListener, error)
	// UpdateEventListener saves everything but Enabled, which only
	// UpdateEventListenerEnabled changes, so neither reverts the other.
	UpdateEventListener(ctx context.Context, eventListener *model.EventListener, change ItemChange) (*model.EventListener, error)
	UpdateEventListenerEnabled(ctx context.Context, id string, enabled bool, updatedAt time.Time) (*model.EventListener, error)
	// UpdateEventListenerWebhookSecret replaces the secret of a webhook
	// listener, which invalidates its previous webhook URL.
	UpdateEventListenerWebhookSecret(ctx context.Context, id string, secret string, updatedAt time.Time) (*model.EventListener, error)
	// EventListenersUpdatedSince includes disabled listeners, except for the
	// zero time.
	EventListenersUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.EventListener, error)
	EnabledEventListenerIDs(ctx context.Context) ([]string, error)
	EnabledScheduledEventListenerIDs(ctx context.Context) ([]string, error)
	// DeleteEventListener keeps the listener's flow in its version history.
	DeleteEventListener(ctx context.Context, id string, userID string) error
}
