package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// EnabledEventListenerCount is how many enabled event listeners of a source
// an app has.
type EnabledEventListenerCount struct {
	AppID  string
	Source model.EventSource
	Count  int
}

type EventListenerStore interface {
	EventListenersByApp(ctx context.Context, appID string) ([]*model.EventListener, error)
	CountEventListenersByAppAndSource(ctx context.Context, appID string, source model.EventSource) (ItemCounts, error)
	UpdateEventListenersLastRunAt(ctx context.Context, lastRunAts map[string]time.Time) error
	EventListener(ctx context.Context, id string) (*model.EventListener, error)
	CreateEventListener(ctx context.Context, eventListener *model.EventListener) (*model.EventListener, error)
	// UpdateEventListener saves everything but Enabled, which only
	// UpdateEventListenerEnabled changes, so neither reverts the other.
	UpdateEventListener(ctx context.Context, eventListener *model.EventListener, change ItemChange) (*model.EventListener, error)
	UpdateEventListenerEnabled(ctx context.Context, id string, enabled bool, updatedAt time.Time) (*model.EventListener, error)
	// AppsWithEnabledEventListenersOver returns the number of enabled event
	// listeners of each app and source that has more than minCount.
	AppsWithEnabledEventListenersOver(ctx context.Context, minCount int) ([]EnabledEventListenerCount, error)
	// DisableNewestEventListeners disables the enabled event listeners of the
	// app and source but the keep oldest ones, and returns how many it
	// disabled.
	DisableNewestEventListeners(ctx context.Context, appID string, source model.EventSource, keep int, updatedAt time.Time) (int, error)
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
