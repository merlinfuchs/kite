package engine

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"gopkg.in/guregu/null.v4"
)

func at(hour, minute, second int) time.Time {
	return time.Date(2026, 9, 25, hour, minute, second, 0, time.UTC)
}

func mustSchedule(t *testing.T, cron string, lastRun null.Time, catchUp bool) *listenerSchedule {
	t.Helper()
	s, err := newListenerSchedule(cron, lastRun, catchUp)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestScheduleDueRunsOnOccurrence(t *testing.T) {
	s := mustSchedule(t, "*/5 * * * *", null.Time{}, false)

	if _, ok := s.due(at(12, 3, 0), time.Second); ok {
		t.Fatal("ran before the first occurrence")
	}
	if _, ok := s.due(at(12, 4, 59), time.Second); ok {
		t.Fatal("ran before the first occurrence")
	}

	occurrence, ok := s.due(at(12, 5, 1), time.Second)
	if !ok || !occurrence.Equal(at(12, 5, 0)) {
		t.Fatalf("due = %s, %v, want 12:05", occurrence, ok)
	}

	if _, ok := s.due(at(12, 5, 2), time.Second); ok {
		t.Fatal("ran the same occurrence twice")
	}
}

func TestScheduleDueSkipsMissedOccurrences(t *testing.T) {
	s := mustSchedule(t, "* * * * * *", null.Time{}, false)

	s.due(at(12, 0, 0), time.Second)

	// The loop stalled for ten seconds, only the latest occurrence runs.
	occurrence, ok := s.due(at(12, 0, 10), time.Second)
	if !ok || !occurrence.Equal(at(12, 0, 10)) {
		t.Fatalf("due = %s, %v, want 12:00:10", occurrence, ok)
	}
	if _, ok := s.due(at(12, 0, 10), time.Second); ok {
		t.Fatal("ran a missed occurrence")
	}
}

func TestScheduleDueEnforcesMinInterval(t *testing.T) {
	s := mustSchedule(t, "* * * * *", null.Time{}, false)

	s.due(at(12, 0, 0), 5*time.Minute)
	if _, ok := s.due(at(12, 1, 0), 5*time.Minute); !ok {
		t.Fatal("first run was skipped")
	}

	for minute := 2; minute < 6; minute++ {
		if _, ok := s.due(at(12, minute, 0), 5*time.Minute); ok {
			t.Fatalf("ran at 12:%02d, within the minimum interval", minute)
		}
	}

	if _, ok := s.due(at(12, 6, 0), 5*time.Minute); !ok {
		t.Fatal("run after the minimum interval was skipped")
	}
}

func TestScheduleDueCatchesUpWithinGracePeriod(t *testing.T) {
	// Last ran at 12:00, the engine was down for the 12:05 and 12:10 runs.
	s := mustSchedule(t, "*/5 * * * *", null.TimeFrom(at(12, 0, 0)), true)

	occurrence, ok := s.due(at(12, 12, 0), time.Second)
	if !ok || !occurrence.Equal(at(12, 10, 0)) {
		t.Fatalf("due = %s, %v, want 12:10", occurrence, ok)
	}
	if _, ok := s.due(at(12, 12, 1), time.Second); ok {
		t.Fatal("caught up on more than one run")
	}
}

func TestScheduleDueDoesNotCatchUpBeyondGracePeriod(t *testing.T) {
	s := mustSchedule(t, "0 * * * *", null.TimeFrom(at(10, 0, 0)), true)

	// 12:00 was missed but is more than the grace period ago.
	if _, ok := s.due(at(12, 30, 0), time.Second); ok {
		t.Fatal("caught up on a run outside the grace period")
	}
}

func TestScheduleDueDoesNotCatchUpWithoutFlag(t *testing.T) {
	// Listeners that are enabled or edited don't catch up.
	s := mustSchedule(t, "*/5 * * * *", null.TimeFrom(at(12, 0, 0)), false)

	if _, ok := s.due(at(12, 7, 0), time.Second); ok {
		t.Fatal("caught up without the catch up flag")
	}
}

func TestScheduleTakeOverKeepsLastRun(t *testing.T) {
	old := mustSchedule(t, "* * * * *", null.Time{}, false)
	old.due(at(12, 0, 0), time.Second)
	if _, ok := old.due(at(12, 1, 0), time.Second); !ok {
		t.Fatal("old schedule didn't run")
	}

	// A reload loaded the listener before its last run was persisted.
	s := mustSchedule(t, "* * * * *", null.Time{}, true)
	s.takeOver(old)

	if !s.lastRun.Equal(at(12, 1, 0)) {
		t.Fatalf("lastRun = %s, want 12:01", s.lastRun)
	}
	if s.catchUp {
		t.Fatal("reloaded listener would catch up")
	}
}

type fakeScheduleListenerStore struct {
	store.EventListenerStore
	listeners []*model.EventListener

	mu       sync.Mutex
	lastRuns map[string]time.Time
}

func (f *fakeScheduleListenerStore) EventListenersUpdatedSince(ctx context.Context, since time.Time) ([]*model.EventListener, error) {
	return f.listeners, nil
}

func (f *fakeScheduleListenerStore) UpdateEventListenersLastRunAt(ctx context.Context, lastRuns map[string]time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, t := range lastRuns {
		f.lastRuns[id] = t
	}
	return nil
}

type fakeLogStore struct {
	store.LogStore
	entries chan model.LogEntry
}

func (f *fakeLogStore) CreateLogEntry(ctx context.Context, entry model.LogEntry) error {
	f.entries <- entry
	return nil
}

type fakeUsageStore struct {
	store.UsageStore
}

func (f *fakeUsageStore) CreateUsageRecord(ctx context.Context, record model.UsageRecord) error {
	return nil
}

type fakeSessions struct{}

func (fakeSessions) AppIDs() []string { return []string{"app"} }

func (fakeSessions) AppSession(ctx context.Context, appID string) (*state.State, error) {
	// Never connects, the flow in the test doesn't call Discord.
	return state.New("Bot test"), nil
}

type fakeFeatures struct {
	minIntervalSeconds int
}

func (f fakeFeatures) AppFeatures(ctx context.Context, appID string) model.Features {
	return model.Features{MinScheduleIntervalSeconds: f.minIntervalSeconds}
}

func (f fakeFeatures) AppFeaturesForApps(ctx context.Context, appIDs []string) (map[string]model.Features, error) {
	res := make(map[string]model.Features, len(appIDs))
	for _, id := range appIDs {
		res[id] = model.Features{MinScheduleIntervalSeconds: f.minIntervalSeconds}
	}
	return res, nil
}

func scheduledListener(cron string) *model.EventListener {
	return &model.EventListener{
		ID:      "listener",
		AppID:   "app",
		Source:  model.EventSourceSchedule,
		Type:    model.EventListenerTypeScheduleCron,
		Enabled: true,
		FlowSource: flow.FlowData{
			Nodes: []flow.FlowNode{
				{ID: "entry", Type: flow.FlowNodeTypeEntryEvent, Data: flow.FlowNodeData{
					EventType:         flow.EventTypeScheduleCron,
					EventScheduleCron: cron,
					Description:       "test",
				}},
				{ID: "log", Type: flow.FlowNodeTypeActionLog, Data: flow.FlowNodeData{
					LogLevel:   provider.LogLevelInfo,
					LogMessage: "ran at {{schedule.time}}",
				}},
			},
			Edges: []flow.FlowEdge{{Source: "entry", Target: "log"}},
		},
	}
}

func newTestScheduler(e *Engine, sessions SessionProvider) *scheduler {
	return &scheduler{
		engine:   e,
		sessions: sessions,
		features: fakeFeatures{minIntervalSeconds: 60},
		lastRuns: make(map[string]time.Time),
	}
}

func newScheduleTestEngine(listeners *fakeScheduleListenerStore, logs *fakeLogStore) *Engine {
	return NewEngine(Env{
		Config:              EngineConfig{ClusterCount: 1},
		EventListenerStore:  listeners,
		LogStore:            logs,
		UsageStore:          &fakeUsageStore{},
		CommandStore:        &fakeCommandStore{},
		PluginInstanceStore: &fakePluginInstanceStore{},
		DeletedEntityStore:  &fakeDeletedEntityStore{},
	})
}

func TestSchedulerRunsScheduledFlow(t *testing.T) {
	listeners := &fakeScheduleListenerStore{
		lastRuns:  make(map[string]time.Time),
		listeners: []*model.EventListener{scheduledListener("* * * * *")},
	}
	logs := &fakeLogStore{entries: make(chan model.LogEntry, 1)}

	e := newScheduleTestEngine(listeners, logs)
	e.populate(context.Background())

	s := newTestScheduler(e, fakeSessions{})

	ctx := context.Background()
	s.tick(ctx, at(12, 0, 30))
	s.tick(ctx, at(12, 1, 0))

	select {
	case entry := <-logs.entries:
		if entry.Message != "ran at 2026-09-25T12:01:00Z" {
			t.Fatalf("log message = %q", entry.Message)
		}
		if entry.EventListenerID.String != "listener" {
			t.Fatalf("log not attributed to the listener: %+v", entry)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("scheduled flow didn't run")
	}

	s.flush(ctx)
	if got := listeners.lastRuns["listener"]; !got.Equal(at(12, 1, 0)) {
		t.Fatalf("persisted last run = %s, want 12:01", got)
	}
}

func TestScheduleDueDropsOccurrenceOlderThanGracePeriod(t *testing.T) {
	s := mustSchedule(t, "0 * * * *", null.Time{}, false)
	s.due(at(11, 59, 0), time.Second)

	// The 12:00 occurrence was never checked for, e.g. because the gateway
	// was down, and is now more than the grace period old.
	if _, ok := s.due(at(12, 10, 0), time.Second); ok {
		t.Fatal("ran an occurrence older than the grace period")
	}
}

func TestScheduleTakeOverKeepsPendingOccurrence(t *testing.T) {
	old := mustSchedule(t, "0 * * * *", null.Time{}, false)
	old.due(at(11, 59, 0), time.Second)

	// Reloaded right before 12:00, first checked just after it.
	s := mustSchedule(t, "0 * * * *", null.Time{}, false)
	s.takeOver(old)

	if occurrence, ok := s.due(at(12, 0, 1), time.Second); !ok || !occurrence.Equal(at(12, 0, 0)) {
		t.Fatalf("due = %s, %v, want 12:00", occurrence, ok)
	}
	if s.running != old.running {
		t.Fatal("running flag isn't shared with the replaced schedule")
	}
}

func TestScheduleTakeOverResetsPendingOccurrenceWhenCronChanged(t *testing.T) {
	old := mustSchedule(t, "0 * * * *", null.Time{}, false)
	old.due(at(11, 59, 0), time.Second)

	s := mustSchedule(t, "30 * * * *", null.Time{}, false)
	s.takeOver(old)

	if _, ok := s.due(at(12, 0, 1), time.Second); ok {
		t.Fatal("ran an occurrence of the old schedule")
	}
}

type flakySessions struct{ up bool }

func (f *flakySessions) AppIDs() []string { return nil }

func (f *flakySessions) AppSession(ctx context.Context, appID string) (*state.State, error) {
	if !f.up {
		return nil, store.ErrNotFound
	}
	return state.New("Bot test"), nil
}

func TestSchedulerWaitsForSession(t *testing.T) {
	listeners := &fakeScheduleListenerStore{
		lastRuns:  make(map[string]time.Time),
		listeners: []*model.EventListener{scheduledListener("* * * * *")},
	}
	logs := &fakeLogStore{entries: make(chan model.LogEntry, 1)}
	e := newScheduleTestEngine(listeners, logs)
	e.populate(context.Background())

	sessions := &flakySessions{}
	s := newTestScheduler(e, sessions)

	ctx := context.Background()
	s.tick(ctx, at(12, 0, 30))
	s.tick(ctx, at(12, 1, 0))

	// The gateway comes up a few seconds after the 12:01 occurrence.
	sessions.up = true
	s.tick(ctx, at(12, 1, 5))

	select {
	case entry := <-logs.entries:
		if entry.Message != "ran at 2026-09-25T12:01:00Z" {
			t.Fatalf("log message = %q", entry.Message)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("occurrence was lost while the gateway was down")
	}
}

func TestPopulateDropsDisabledListener(t *testing.T) {
	listener := scheduledListener("* * * * *")
	listeners := &fakeScheduleListenerStore{
		lastRuns:  make(map[string]time.Time),
		listeners: []*model.EventListener{listener},
	}
	e := newScheduleTestEngine(listeners, &fakeLogStore{})
	e.populate(context.Background())

	if got := len(e.scheduledEventListeners()); got != 1 {
		t.Fatalf("scheduled listeners = %d, want 1", got)
	}

	disabled := *listener
	disabled.Enabled = false
	listeners.listeners = []*model.EventListener{&disabled}
	e.populate(context.Background())

	if got := len(e.scheduledEventListeners()); got != 0 {
		t.Fatalf("scheduled listeners after disabling = %d, want 0", got)
	}
}

func TestPopulateDropsDeletedScheduledListener(t *testing.T) {
	listeners := &fakeScheduleListenerStore{
		lastRuns:  make(map[string]time.Time),
		listeners: []*model.EventListener{scheduledListener("* * * * *")},
	}
	e := newScheduleTestEngine(listeners, &fakeLogStore{})
	e.populate(context.Background())

	// Deleted rows never show up as updated again, only their tombstone does.
	listeners.listeners = nil
	e.env.DeletedEntityStore = &fakeDeletedEntityStore{deleted: []*model.DeletedEntity{
		{ID: "listener", Type: model.DeletedEntityTypeEventListener, AppID: "app"},
	}}
	e.populate(context.Background())

	if got := len(e.scheduledEventListeners()); got != 0 {
		t.Fatalf("scheduled listeners after deleting = %d, want 0", got)
	}
}

func TestScheduleRetiredByReloadDoesNotRun(t *testing.T) {
	old := mustSchedule(t, "0 * * * *", null.Time{}, false)
	old.due(at(11, 59, 0), time.Second)

	s := mustSchedule(t, "0 * * * *", null.Time{}, false)
	s.takeOver(old)

	// A tick that still held the old listener.
	if _, ok := old.due(at(12, 0, 1), time.Second); ok {
		t.Fatal("replaced schedule ran")
	}
	if _, ok := s.due(at(12, 0, 1), time.Second); !ok {
		t.Fatal("reloaded schedule didn't run")
	}
}

func TestScheduleDueSkipsLongBacklogQuickly(t *testing.T) {
	s := mustSchedule(t, "* * * * * *", null.Time{}, false)
	s.due(at(12, 0, 0), time.Second)

	// A week of missed per-second occurrences, e.g. while the gateway was
	// down, only the grace period is stepped through.
	start := time.Now()
	s.due(at(12, 0, 0).Add(7*24*time.Hour), time.Second)
	if time.Since(start) > time.Second {
		t.Fatalf("due took %s for a long backlog", time.Since(start))
	}
}
