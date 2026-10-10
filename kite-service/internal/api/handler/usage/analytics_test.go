package usage

import (
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

func TestAnalyticsRangeFor(t *testing.T) {
	now := time.Date(2026, 9, 28, 14, 35, 0, 0, time.UTC)

	tests := []struct {
		name    string
		bucket  analyticsBucket
		start   time.Time
		end     time.Time
		buckets int
	}{
		{"1d", analyticsBucketHour, time.Date(2026, 9, 27, 15, 0, 0, 0, time.UTC), time.Date(2026, 9, 28, 15, 0, 0, 0, time.UTC), 24},
		{"1w", analyticsBucketDay, time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC), 7},
		{"1m", analyticsBucketDay, time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC), 30},
		{"1y", analyticsBucketMonth, time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), 12},
	}

	for _, tt := range tests {
		r, ok := analyticsRangeFor(tt.name, now)
		if !ok {
			t.Fatalf("%s: range not accepted", tt.name)
		}
		if r.bucket != tt.bucket || !r.start.Equal(tt.start) || !r.end.Equal(tt.end) {
			t.Errorf("%s: got %s %s-%s, want %s %s-%s", tt.name, r.bucket, r.start, r.end, tt.bucket, tt.start, tt.end)
		}
		if !r.hasPrevious || !r.previousStart.Before(r.start) {
			t.Errorf("%s: expected a previous period before the start", tt.name)
		}

		series, _ := analyticsSeries(nil, r.bucket, r.start, r.end)
		if len(series) != tt.buckets {
			t.Errorf("%s: got %d buckets, want %d", tt.name, len(series), tt.buckets)
		}
	}

	r, ok := analyticsRangeFor("all", now)
	if !ok || r.hasPrevious || !r.allTime {
		t.Errorf("all: got %+v", r)
	}

	for _, name := range []string{"", "2d", "ALL", "1month"} {
		if _, ok := analyticsRangeFor(name, now); ok {
			t.Errorf("%q: expected range to be rejected", name)
		}
	}
}

func TestAnalyticsSeries(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)

	series, totals := analyticsSeries([]model.UsageAnalyticsBucket{
		{Time: start, Type: model.UsageRecordTypeCommandFlowExecution, Executions: 3, CreditsUsed: 6},
		{Time: start, Type: model.UsageRecordTypeMessageFlowExecution, Executions: 1, CreditsUsed: 1},
		{Time: start.AddDate(0, 0, 2), Type: model.UsageRecordTypeEventListenerFlowExecution, Executions: 5, CreditsUsed: 10},
	}, analyticsBucketDay, start, end)

	if len(series) != 3 {
		t.Fatalf("got %d entries, want 3", len(series))
	}
	if series[0].CreditsUsed != 7 || series[0].CommandExecutions != 3 || series[0].MessageExecutions != 1 {
		t.Errorf("first day: got %+v", series[0])
	}
	if series[1].CreditsUsed != 0 {
		t.Errorf("empty day: got %+v", series[1])
	}
	if series[2].EventListenerExecutions != 5 {
		t.Errorf("last day: got %+v", series[2])
	}

	if totals.Executions != 9 || totals.CreditsUsed != 17 || totals.CommandExecutions != 3 ||
		totals.EventListenerExecutions != 5 || totals.MessageExecutions != 1 || totals.CommandCreditsUsed != 6 {
		t.Errorf("totals: got %+v", totals)
	}
}
