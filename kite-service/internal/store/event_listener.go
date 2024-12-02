package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type EventListenerStore interface {
	EventListenersByApp(ctx context.Context, appID string) ([]*model.EventListener, error)
	CountEventListenersByApp(ctx context.Context, appID string) (int, error)
	EventListener(ctx context.Context, id string) (*model.EventListener, error)
	CreateEventListener(ctx context.Context, eventListener *model.EventListener) (*model.EventListener, error)
	UpdateEventListener(ctx context.Context, eventListener *model.EventListener) (*model.EventListener, error)
	EnabledEventListenersUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.EventListener, error)
	EnabledEventListenerIDs(ctx context.Context) ([]string, error)
	DeleteEventListener(ctx context.Context, id string) error
}
