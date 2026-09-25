package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type EventListenerStore interface {
	EventListenersByApp(ctx context.Context, appID string) ([]*model.EventListener, error)
	CountEventListenersByAppAndSource(ctx context.Context, appID string, source model.EventSource) (int, error)
	UpdateEventListenersLastRunAt(ctx context.Context, lastRunAts map[string]time.Time) error
	EventListener(ctx context.Context, id string) (*model.EventListener, error)
	CreateEventListener(ctx context.Context, eventListener *model.EventListener) (*model.EventListener, error)
	UpdateEventListener(ctx context.Context, eventListener *model.EventListener) (*model.EventListener, error)
	// EventListenersUpdatedSince includes disabled listeners, except for the
	// zero time.
	EventListenersUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.EventListener, error)
	EnabledEventListenerIDs(ctx context.Context) ([]string, error)
	EnabledScheduledEventListenerIDs(ctx context.Context) ([]string, error)
	DeleteEventListener(ctx context.Context, id string) error
}
