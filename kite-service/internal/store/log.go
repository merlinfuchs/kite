package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type LogStore interface {
	CreateLogEntry(ctx context.Context, entry model.LogEntry) error
	LogEntriesByApp(ctx context.Context, appID string, beforeID int64, limit int) ([]*model.LogEntry, error)
	LogEntriesByCommand(ctx context.Context, appID string, commandID string, beforeID int64, limit int) ([]*model.LogEntry, error)
	LogEntriesByEvent(ctx context.Context, appID string, eventID string, beforeID int64, limit int) ([]*model.LogEntry, error)
	LogEntriesByMessage(ctx context.Context, appID string, messageID string, beforeID int64, limit int) ([]*model.LogEntry, error)
	LogSummary(ctx context.Context, appID string, start time.Time, end time.Time) (*model.LogSummary, error)
}
