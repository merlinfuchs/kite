package wire

import "time"

type UsageCreditsGetResponse struct {
	CreditsUsed int `json:"credits_used"`
}

type UsageByDayListResponse []*UsageByDayEntry

type UsageByDayEntry struct {
	Date        time.Time `json:"date"`
	CreditsUsed int       `json:"credits_used"`
}

type UsageByTypeListResponse []*UsageByTypeEntry

type UsageByTypeEntry struct {
	Type        string `json:"type"`
	CreditsUsed int    `json:"credits_used"`
}

type UsageAnalyticsTotals struct {
	Executions               int64 `json:"executions"`
	CreditsUsed              int64 `json:"credits_used"`
	CommandExecutions        int64 `json:"command_executions"`
	CommandCreditsUsed       int64 `json:"command_credits_used"`
	EventListenerExecutions  int64 `json:"event_listener_executions"`
	EventListenerCreditsUsed int64 `json:"event_listener_credits_used"`
	MessageExecutions        int64 `json:"message_executions"`
	MessageCreditsUsed       int64 `json:"message_credits_used"`
}

type UsageAnalyticsSeriesEntry struct {
	Time                    time.Time `json:"time"`
	CreditsUsed             int64     `json:"credits_used"`
	CommandExecutions       int64     `json:"command_executions"`
	EventListenerExecutions int64     `json:"event_listener_executions"`
	MessageExecutions       int64     `json:"message_executions"`
}

type UsageAnalyticsSourceEntry struct {
	ID          string `json:"id"`
	Executions  int64  `json:"executions"`
	CreditsUsed int64  `json:"credits_used"`
}

type UsageAnalyticsLogs struct {
	Errors   int64 `json:"errors"`
	Warnings int64 `json:"warnings"`
	// Partial is true when the range reaches further back than logs are kept.
	Partial bool `json:"partial"`
}

type UsageAnalyticsGetResponse struct {
	Range string `json:"range"`
	// Bucket is the size of each series entry: hour, day or month.
	Bucket  string               `json:"bucket"`
	StartAt time.Time            `json:"start_at"`
	EndAt   time.Time            `json:"end_at"`
	Totals  UsageAnalyticsTotals `json:"totals"`
	// PreviousTotals covers the period of the same length right before, null for all time.
	PreviousTotals    *UsageAnalyticsTotals        `json:"previous_totals"`
	Series            []*UsageAnalyticsSeriesEntry `json:"series"`
	TopCommands       []*UsageAnalyticsSourceEntry `json:"top_commands"`
	TopEventListeners []*UsageAnalyticsSourceEntry `json:"top_event_listeners"`
	TopMessages       []*UsageAnalyticsSourceEntry `json:"top_messages"`
	Logs              UsageAnalyticsLogs           `json:"logs"`
}
