package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type NotificationStore interface {
	// CreateNotification reports whether the notification was created, which
	// it isn't if one with the same type and dedupe key exists.
	CreateNotification(ctx context.Context, notification *model.Notification) (bool, error)
	// DeleteNotificationsBefore deletes up to batchSize notifications created
	// before createdBefore.
	DeleteNotificationsBefore(ctx context.Context, createdBefore time.Time, batchSize int) (int64, error)
}
