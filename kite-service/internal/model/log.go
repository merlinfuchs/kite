package model

import "time"

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogEntry struct {
	ID        int64     `json:"id"`
	AppID     string    `json:"app_id"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type LogSummary struct {
	TotalEntries  int64 `json:"total_entries"`
	TotalErrors   int64 `json:"total_errors"`
	TotalWarnings int64 `json:"total_warnings"`
	TotalInfos    int64 `json:"total_infos"`
	TotalDebugs   int64 `json:"total_debugs"`
}
