package wire

import (
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type LogEntry struct {
	ID              int64       `json:"id"`
	Message         string      `json:"message"`
	Level           string      `json:"level"`
	CommandID       null.String `json:"command_id"`
	EventListenerID null.String `json:"event_listener_id"`
	MessageID       null.String `json:"message_id"`
	CreatedAt       time.Time   `json:"created_at"`
}

type LogEntryListResponse = []*LogEntry

type LogSummary struct {
	TotalEntries  int64 `json:"total_entries"`
	TotalErrors   int64 `json:"total_errors"`
	TotalWarnings int64 `json:"total_warnings"`
	TotalInfos    int64 `json:"total_infos"`
	TotalDebugs   int64 `json:"total_debugs"`
}

type LogSummaryGetResponse = LogSummary

func LogEntryToWire(entry *model.LogEntry) *LogEntry {
	if entry == nil {
		return nil
	}

	return &LogEntry{
		ID:              entry.ID,
		Message:         entry.Message,
		Level:           string(entry.Level),
		CommandID:       entry.CommandID,
		EventListenerID: entry.EventListenerID,
		MessageID:       entry.MessageID,
		CreatedAt:       entry.CreatedAt,
	}
}

func LogSummaryToWire(summary *model.LogSummary) *LogSummary {
	if summary == nil {
		return nil
	}

	return &LogSummary{
		TotalEntries:  summary.TotalEntries,
		TotalErrors:   summary.TotalErrors,
		TotalWarnings: summary.TotalWarnings,
		TotalInfos:    summary.TotalInfos,
		TotalDebugs:   summary.TotalDebugs,
	}
}
