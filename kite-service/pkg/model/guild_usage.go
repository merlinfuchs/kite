package model

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
)

type GuildUsageEntry struct {
	ID                      uint64
	GuildID                 distype.Snowflake
	TotalEventCount         int
	SuccessEventCount       int
	TotalEventExecutionTime time.Duration
	AvgEventExecutionTime   time.Duration
	TotalEventTotalTime     time.Duration
	AvgEventTotalTime       time.Duration
	TotalCallCount          int
	SuccessCallCount        int
	TotalCallTotalTime      time.Duration
	AvgCallTotalTime        time.Duration
	PeriodStartsAt          time.Time
	PeriodEndsAt            time.Time
}

type GuildUsageSummary struct {
	TotalEventCount         int
	SuccessEventCount       int
	TotalEventExecutionTime time.Duration
	AvgEventExecutionTime   time.Duration
	TotalEventTotalTime     time.Duration
	AvgEventTotalTime       time.Duration
	TotalCallCount          int
	SuccessCallCount        int
	TotalCallTotalTime      time.Duration
	AvgCallTotalTime        time.Duration
}

type GuildUsageAndLimits struct {
	Usage  GuildUsageSummary
	Limits GuildEntitlementResolved
}
