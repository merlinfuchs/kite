package usage

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/store"
)

type fakeRecord struct {
	appID     string
	credits   int
	createdAt time.Time
}

type fakeUsageStore struct {
	store.UsageStore
	records []fakeRecord
	queries int
}

func (f *fakeUsageStore) AllUsageCreditsUsedBetween(ctx context.Context, start time.Time, end time.Time) (map[string]int, error) {
	f.queries++
	res := make(map[string]int)
	for _, r := range f.records {
		if !r.createdAt.Before(start) && r.createdAt.Before(end) {
			res[r.appID] += r.credits
		}
	}
	return res, nil
}

func (f *fakeUsageStore) add(appID string, credits int, createdAt time.Time) {
	f.records = append(f.records, fakeRecord{appID, credits, createdAt})
}

func (f *fakeUsageStore) totalsFor(month time.Time) map[string]int {
	start, end := startAndEndOfMonth(month)
	res, _ := f.AllUsageCreditsUsedBetween(context.Background(), start, end)
	return res
}

func assertCredits(t *testing.T, got, want map[string]int) {
	t.Helper()
	for appID, w := range want {
		if got[appID] != w {
			t.Errorf("app %s: credits = %d, want %d", appID, got[appID], w)
		}
	}
	for appID, g := range got {
		if _, ok := want[appID]; !ok && g != 0 {
			t.Errorf("app %s: credits = %d, want none", appID, g)
		}
	}
}

// Sweeping incrementally must always agree with summing the whole month,
// including rows exactly on the settled boundary and rows inserted late.
func TestCreditsUsedThisMonthMatchesFullSum(t *testing.T) {
	ctx := context.Background()
	fs := &fakeUsageStore{}
	m := &UsageManager{usageStore: fs}

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	fs.add("a", 5, time.Date(2026, 8, 31, 23, 59, 0, 0, time.UTC))
	fs.add("a", 10, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	fs.add("b", 3, now.Add(-time.Hour))
	fs.add("a", 1, now.Add(-time.Minute))

	for i := range 10 {
		got, err := m.creditsUsedThisMonth(ctx, now)
		if err != nil {
			t.Fatal(err)
		}
		assertCredits(t, got, fs.totalsFor(now))

		// A row landing exactly on the next settled boundary, and one stamped
		// a few minutes back whose insert only now committed.
		next := now.Add(time.Minute)
		fs.add("a", 2, next.Add(-usageSettleDelay))
		fs.add("c", 7, now.Add(-usageSettleDelay+30*time.Second))
		fs.add("b", i, next)
		now = next
	}
}

func TestCreditsUsedThisMonthResetsOnNewMonth(t *testing.T) {
	ctx := context.Background()
	fs := &fakeUsageStore{}
	m := &UsageManager{usageStore: fs}

	fs.add("a", 10, time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC))
	if _, err := m.creditsUsedThisMonth(ctx, time.Date(2026, 9, 30, 23, 59, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	october := time.Date(2026, 10, 1, 0, 1, 0, 0, time.UTC)
	fs.add("a", 4, october.Add(-30*time.Second))
	got, err := m.creditsUsedThisMonth(ctx, october)
	if err != nil {
		t.Fatal(err)
	}
	assertCredits(t, got, map[string]int{"a": 4})
}

// The point of the settled totals: after the first sweep, later sweeps only
// read the slice that settled since plus the recent tail.
func TestCreditsUsedThisMonthOnlyReadsNewRanges(t *testing.T) {
	ctx := context.Background()
	fs := &fakeUsageStore{}
	m := &UsageManager{usageStore: fs}

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	if _, err := m.creditsUsedThisMonth(ctx, now); err != nil {
		t.Fatal(err)
	}
	if !m.settledUntil.Equal(now.Add(-usageSettleDelay)) {
		t.Fatalf("settledUntil = %v, want %v", m.settledUntil, now.Add(-usageSettleDelay))
	}

	fs.queries = 0
	if _, err := m.creditsUsedThisMonth(ctx, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if fs.queries != 2 {
		t.Errorf("queries = %d, want 2 (newly settled slice + recent tail)", fs.queries)
	}
}
