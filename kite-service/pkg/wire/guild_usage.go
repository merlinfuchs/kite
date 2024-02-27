package wire

import "github.com/merlinfuchs/kite/kite-service/pkg/model"

// GuildUsageSummary represents a summary of the usage of a guild
// All times are in milliseconds
type GuildUsageSummary struct {
	TotalEventCount         int     `json:"total_event_count"`
	SuccessEventCount       int     `json:"success_event_count"`
	TotalEventExecutionTime float32 `json:"total_event_execution_time"`
	AvgEventExecutionTime   float32 `json:"avg_event_execution_time"`
	TotalEventTotalTime     float32 `json:"total_event_total_time"`
	AvgEventTotalTime       float32 `json:"avg_event_total_time"`
	TotalCallCount          int     `json:"total_call_count"`
	SuccessCallCount        int     `json:"success_call_count"`
	TotalCallTotalTime      float32 `json:"total_call_total_time"`
	AvgCallTotalTime        float32 `json:"avg_call_total_time"`
}

type GuildUsageSummaryGetResponse APIResponse[GuildUsageSummary]

func GuildUsageSummaryToWire(d *model.GuildUsageSummary) GuildUsageSummary {
	return GuildUsageSummary{
		TotalEventCount:         d.TotalEventCount,
		SuccessEventCount:       d.SuccessEventCount,
		TotalEventExecutionTime: float32(d.TotalEventExecutionTime.Microseconds()) / 1000,
		AvgEventExecutionTime:   float32(d.AvgEventExecutionTime.Microseconds()) / 1000,
		TotalEventTotalTime:     float32(d.TotalEventTotalTime.Microseconds()) / 1000,
		AvgEventTotalTime:       float32(d.AvgEventTotalTime.Microseconds()) / 1000,
		TotalCallCount:          d.TotalCallCount,
		SuccessCallCount:        d.SuccessCallCount,
		TotalCallTotalTime:      float32(d.TotalCallTotalTime.Microseconds()) / 1000,
		AvgCallTotalTime:        float32(d.AvgCallTotalTime.Microseconds()) / 1000,
	}
}
