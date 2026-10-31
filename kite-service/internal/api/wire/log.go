package wire

import (
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type LogEntry struct {
	ID              int64       `json:"id"`
	Message         string      `json:"message"`
	Level           string      `json:"level"`
	CommandID       null.String `json:"command_id"`
	EventListenerID null.String `json:"event_listener_id"`
	MessageID       null.String `json:"message_id"`
	// The block a flow error happened in.
	NodeID    null.String      `json:"node_id"`
	NodeType  null.String      `json:"node_type"`
	ErrorCode null.String      `json:"error_code"`
	Details   *LogEntryDetails `json:"details,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

type LogEntryDetails struct {
	DiscordStatus int    `json:"discord_status,omitempty"`
	DiscordCode   int    `json:"discord_code,omitempty"`
	Integration   string `json:"integration,omitempty"`
	HTTPStatus    int    `json:"http_status,omitempty"`
	// The setting of the block that was invalid.
	Field string `json:"field,omitempty"`
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
		NodeID:          entry.NodeID,
		NodeType:        entry.NodeType,
		ErrorCode:       entry.ErrorCode,
		Details:         logEntryDetailsToWire(entry.Details),
		CreatedAt:       entry.CreatedAt,
	}
}

func logEntryDetailsToWire(details *flow.ErrorDetails) *LogEntryDetails {
	if details == nil {
		return nil
	}

	return &LogEntryDetails{
		DiscordStatus: details.DiscordStatus,
		DiscordCode:   details.DiscordCode,
		Integration:   details.Integration,
		HTTPStatus:    details.HTTPStatus,
		Field:         details.Field,
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
