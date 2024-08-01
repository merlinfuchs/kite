package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type LogStore interface {
	CreateLogEntry(ctx context.Context, entry model.LogEntry) error
	LogEntriesByApp(ctx context.Context, appID string, beforeID int64, limit int) ([]*model.LogEntry, error)
}
