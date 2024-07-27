package wire

import (
	"time"

	"github.com/kitecloud/kite/kite-common/model"
)

type LogEntry struct {
	ID        int64     `json:"id"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
	CreatedAt time.Time `json:"created_at"`
}

type LogEntryListResponse = []*LogEntry

func LogEntryToWire(entry *model.LogEntry) *LogEntry {
	if entry == nil {
		return nil
	}

	return &LogEntry{
		ID:        entry.ID,
		Message:   entry.Message,
		Level:     string(entry.Level),
		CreatedAt: entry.CreatedAt,
	}
}
