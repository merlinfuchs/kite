package usage

import (
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	usagecore "github.com/kitecloud/kite/kite-service/internal/core/usage"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

const analyticsTopSourcesPerType = 5

type analyticsBucket string

const (
	analyticsBucketHour  analyticsBucket = "hour"
	analyticsBucketDay   analyticsBucket = "day"
	analyticsBucketMonth analyticsBucket = "month"
)

func (b analyticsBucket) next(t time.Time) time.Time {
	switch b {
	case analyticsBucketHour:
		return t.Add(time.Hour)
	case analyticsBucketDay:
		return t.AddDate(0, 0, 1)
	default:
		return t.AddDate(0, 1, 0)
	}
}

func (b analyticsBucket) truncate(t time.Time) time.Time {
	switch b {
	case analyticsBucketHour:
		return t.Truncate(time.Hour)
	case analyticsBucketDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	default:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
}

type analyticsRange struct {
	bucket analyticsBucket
	start  time.Time
	end    time.Time
	// hasPrevious is false for all time, there is nothing before it.
	hasPrevious   bool
	previousStart time.Time
	// allTime starts the series at the first bucket with data instead of start.
	allTime bool
}

// analyticsRangeFor returns the range ending with the bucket that contains now.
func analyticsRangeFor(name string, now time.Time) (analyticsRange, bool) {
	now = now.UTC()

	var r analyticsRange
	switch name {
	case "1d":
		r.bucket = analyticsBucketHour
		r.end = r.bucket.next(r.bucket.truncate(now))
		r.start = r.end.Add(-24 * time.Hour)
		r.previousStart = r.start.Add(-24 * time.Hour)
	case "1w", "1m":
		days := 7
		if name == "1m" {
			days = 30
		}
		r.bucket = analyticsBucketDay
		r.end = r.bucket.next(r.bucket.truncate(now))
		r.start = r.end.AddDate(0, 0, -days)
		r.previousStart = r.start.AddDate(0, 0, -days)
	case "1y":
		r.bucket = analyticsBucketMonth
		r.end = r.bucket.next(r.bucket.truncate(now))
		r.start = r.end.AddDate(0, -12, 0)
		r.previousStart = r.start.AddDate(0, -12, 0)
	case "all":
		r.bucket = analyticsBucketMonth
		r.end = r.bucket.next(r.bucket.truncate(now))
		r.start = time.Unix(0, 0).UTC()
		r.allTime = true
		return r, true
	default:
		return r, false
	}

	r.hasPrevious = true
	return r, true
}

func (h *UsageHandler) HandleUsageAnalyticsGet(c *handler.Context) (*wire.UsageAnalyticsGetResponse, error) {
	rangeName := c.Query("range")
	if rangeName == "" {
		rangeName = "1m"
	}

	now := time.Now().UTC()

	r, ok := analyticsRangeFor(rangeName, now)
	if !ok {
		return nil, handler.ErrBadRequest("invalid_range", "range must be one of 1d, 1w, 1m, 1y or all")
	}

	buckets, err := h.usageStore.UsageAnalyticsSeriesBetween(c.Context(), c.App.ID, r.start, r.end, string(r.bucket))
	if err != nil {
		return nil, fmt.Errorf("failed to get usage analytics series: %w", err)
	}

	res := &wire.UsageAnalyticsGetResponse{
		Range:  rangeName,
		Bucket: string(r.bucket),
		EndAt:  r.end,
	}

	seriesStart := r.start
	if r.allTime {
		// Starts at the first month with usage, or the current month if there is none.
		seriesStart = r.bucket.truncate(now)
		if len(buckets) > 0 && buckets[0].Time.Before(seriesStart) {
			seriesStart = r.bucket.truncate(buckets[0].Time)
		}
	}
	res.StartAt = seriesStart
	res.Series, res.Totals = analyticsSeries(buckets, r.bucket, seriesStart, r.end)

	if r.hasPrevious {
		totals, err := h.usageStore.UsageAnalyticsTotalsBetween(c.Context(), c.App.ID, r.previousStart, r.start)
		if err != nil {
			return nil, fmt.Errorf("failed to get previous usage analytics totals: %w", err)
		}

		previous := analyticsTotals(totals)
		res.PreviousTotals = &previous
	}

	sources, err := h.usageStore.UsageAnalyticsTopSourcesBetween(c.Context(), c.App.ID, r.start, r.end, analyticsTopSourcesPerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage analytics top sources: %w", err)
	}

	res.TopCommands = []*wire.UsageAnalyticsSourceEntry{}
	res.TopEventListeners = []*wire.UsageAnalyticsSourceEntry{}
	res.TopMessages = []*wire.UsageAnalyticsSourceEntry{}
	for _, source := range sources {
		entry := &wire.UsageAnalyticsSourceEntry{
			ID:          source.SourceID,
			Executions:  source.Executions,
			CreditsUsed: source.CreditsUsed,
		}

		switch source.Type {
		case model.UsageRecordTypeCommandFlowExecution:
			res.TopCommands = append(res.TopCommands, entry)
		case model.UsageRecordTypeEventListenerFlowExecution:
			res.TopEventListeners = append(res.TopEventListeners, entry)
		case model.UsageRecordTypeMessageFlowExecution:
			res.TopMessages = append(res.TopMessages, entry)
		}
	}

	// Logs aren't rolled up, so longer ranges only cover what is still kept.
	logStart := r.start
	if logsKeptSince := now.Add(-usagecore.LogEntryExpiry); logStart.Before(logsKeptSince) {
		logStart = logsKeptSince
		res.Logs.Partial = true
	}

	summary, err := h.logStore.LogSummary(c.Context(), c.App.ID, logStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get log summary: %w", err)
	}
	if summary != nil {
		res.Logs.Errors = summary.TotalErrors
		res.Logs.Warnings = summary.TotalWarnings
	}

	return res, nil
}

// analyticsSeries returns one entry per bucket between start and end, empty
// buckets included, and the totals across all of them.
func analyticsSeries(
	buckets []model.UsageAnalyticsBucket,
	bucket analyticsBucket,
	start time.Time,
	end time.Time,
) ([]*wire.UsageAnalyticsSeriesEntry, wire.UsageAnalyticsTotals) {
	series := []*wire.UsageAnalyticsSeriesEntry{}
	byTime := make(map[time.Time]*wire.UsageAnalyticsSeriesEntry)
	for t := start; t.Before(end); t = bucket.next(t) {
		entry := &wire.UsageAnalyticsSeriesEntry{Time: t}
		series = append(series, entry)
		byTime[t] = entry
	}

	var totals wire.UsageAnalyticsTotals
	for _, b := range buckets {
		addAnalyticsTotal(&totals, b.Type, b.Executions, b.CreditsUsed)

		entry, ok := byTime[bucket.truncate(b.Time.UTC())]
		if !ok {
			continue
		}

		entry.CreditsUsed += b.CreditsUsed
		switch b.Type {
		case model.UsageRecordTypeCommandFlowExecution:
			entry.CommandExecutions += b.Executions
		case model.UsageRecordTypeEventListenerFlowExecution:
			entry.EventListenerExecutions += b.Executions
		case model.UsageRecordTypeMessageFlowExecution:
			entry.MessageExecutions += b.Executions
		}
	}

	return series, totals
}

func analyticsTotals(entries []model.UsageAnalyticsTotal) wire.UsageAnalyticsTotals {
	var totals wire.UsageAnalyticsTotals
	for _, e := range entries {
		addAnalyticsTotal(&totals, e.Type, e.Executions, e.CreditsUsed)
	}
	return totals
}

func addAnalyticsTotal(totals *wire.UsageAnalyticsTotals, t model.UsageRecordType, executions int64, creditsUsed int64) {
	totals.Executions += executions
	totals.CreditsUsed += creditsUsed

	switch t {
	case model.UsageRecordTypeCommandFlowExecution:
		totals.CommandExecutions += executions
		totals.CommandCreditsUsed += creditsUsed
	case model.UsageRecordTypeEventListenerFlowExecution:
		totals.EventListenerExecutions += executions
		totals.EventListenerCreditsUsed += creditsUsed
	case model.UsageRecordTypeMessageFlowExecution:
		totals.MessageExecutions += executions
		totals.MessageCreditsUsed += creditsUsed
	}
}
