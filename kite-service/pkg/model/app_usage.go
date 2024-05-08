package model

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
)

type AppUsageEntry struct {
	ID                      uint64
	AppID                   distype.Snowflake
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

type AppUsageSummary struct {
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

type AppUsageAndLimits struct {
	AppUsageSummary
	Limits AppEntitlementResolved
}
