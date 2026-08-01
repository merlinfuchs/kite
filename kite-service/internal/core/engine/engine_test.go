package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// The fakes embed their store interface so any method the engine does not call
// panics rather than silently returning a zero value.

type fakeCommandStore struct {
	store.CommandStore
	err        error
	calledWith []time.Time
}

func (f *fakeCommandStore) EnabledCommandsUpdatedSince(ctx context.Context, since time.Time) ([]*model.Command, error) {
	f.calledWith = append(f.calledWith, since)
	return nil, f.err
}

type fakeEventListenerStore struct {
	store.EventListenerStore
	err error
}

func (f *fakeEventListenerStore) EnabledEventListenersUpdatedSince(ctx context.Context, since time.Time) ([]*model.EventListener, error) {
	return nil, f.err
}

type fakePluginInstanceStore struct {
	store.PluginInstanceStore
	err error
}

func (f *fakePluginInstanceStore) EnabledPluginInstancesUpdatedSince(ctx context.Context, since time.Time) ([]*model.PluginInstance, error) {
	return nil, f.err
}

func newTestEngine(commands *fakeCommandStore, listeners *fakeEventListenerStore, plugins *fakePluginInstanceStore) *Engine {
	return NewEngine(Env{
		Config: EngineConfig{
			ClusterCount:    1,
			ClusterIndex:    0,
			PopulateOverlap: 5 * time.Second,
		},
		CommandStore:        commands,
		EventListenerStore:  listeners,
		PluginInstanceStore: plugins,
	})
}

// On success the cursor moves forward, but rewound by the overlap, so rows
// committed while the queries were in flight are re-read next time instead of
// being skipped forever.
func TestPopulateAdvancesCursorWithOverlap(t *testing.T) {
	commands := &fakeCommandStore{}
	e := newTestEngine(commands, &fakeEventListenerStore{}, &fakePluginInstanceStore{})

	before := time.Now().UTC()
	e.populate(context.Background())
	after := time.Now().UTC()

	if e.lastUpdate.IsZero() {
		t.Fatal("cursor was not advanced after a successful poll")
	}

	// The cursor should sit roughly one overlap behind the moment the poll
	// started, and never ahead of it.
	if !e.lastUpdate.Before(before) {
		t.Errorf("cursor = %v, want at least %v earlier than poll start %v",
			e.lastUpdate, 5*time.Second, before)
	}

	earliestAcceptable := before.Add(-5*time.Second - time.Second)
	if e.lastUpdate.Before(earliestAcceptable) {
		t.Errorf("cursor = %v, rewound further than the configured overlap (poll ran %v..%v)",
			e.lastUpdate, before, after)
	}
}

// A failed query must leave the cursor alone. Advancing it past a window that
// was never read means those rows are never picked up again.
func TestPopulateDoesNotAdvanceCursorOnError(t *testing.T) {
	commands := &fakeCommandStore{err: errors.New("connection refused")}
	e := newTestEngine(commands, &fakeEventListenerStore{}, &fakePluginInstanceStore{})

	e.populate(context.Background())

	if !e.lastUpdate.IsZero() {
		t.Errorf("cursor advanced to %v despite a query failure, want it left at zero", e.lastUpdate)
	}
}

// The window a failed poll covered must be retried verbatim by the next poll.
func TestPopulateRetriesSameWindowAfterError(t *testing.T) {
	commands := &fakeCommandStore{err: errors.New("connection refused")}
	e := newTestEngine(commands, &fakeEventListenerStore{}, &fakePluginInstanceStore{})

	e.populate(context.Background())
	e.populate(context.Background())

	if len(commands.calledWith) != 2 {
		t.Fatalf("store queried %d times, want 2", len(commands.calledWith))
	}
	if !commands.calledWith[0].Equal(commands.calledWith[1]) {
		t.Errorf("retry queried from %v, want the same cursor as the failed poll (%v)",
			commands.calledWith[1], commands.calledWith[0])
	}
}

// A partial failure must not advance the cursor either, even though the other
// two queries succeeded: the failed one still has an unread window.
func TestPopulatePartialFailureDoesNotAdvanceCursor(t *testing.T) {
	e := newTestEngine(
		&fakeCommandStore{},
		&fakeEventListenerStore{err: errors.New("connection refused")},
		&fakePluginInstanceStore{},
	)

	e.populate(context.Background())

	if !e.lastUpdate.IsZero() {
		t.Errorf("cursor advanced to %v despite one of three queries failing", e.lastUpdate)
	}
}

func TestIDLookupSet(t *testing.T) {
	set := idLookupSet([]string{"a", "b", "a"})

	if len(set) != 2 {
		t.Errorf("idLookupSet produced %d entries, want 2 after dedup", len(set))
	}
	for _, want := range []string{"a", "b"} {
		if _, ok := set[want]; !ok {
			t.Errorf("idLookupSet is missing %q", want)
		}
	}
	if _, ok := set["c"]; ok {
		t.Error("idLookupSet contains an id that was not passed in")
	}
}

func TestIDLookupSetEmpty(t *testing.T) {
	// A sweep that finds no enabled entities must still produce a usable set,
	// so the per-app calls remove everything rather than panicking.
	set := idLookupSet(nil)
	if set == nil {
		t.Fatal("idLookupSet(nil) returned a nil map")
	}
	if len(set) != 0 {
		t.Errorf("idLookupSet(nil) produced %d entries, want 0", len(set))
	}
}

// Reproduces the production shape of a dangling sweep: many apps in the
// registry, all tested against the system-wide set of enabled entities.
//
// The set is now built once per sweep. Building it per app (the previous
// behaviour) made the sweep O(apps x entities) and, at ~5k apps against a
// system-wide set, accounted for 74% of the process's CPU.
func benchmarkSweep(b *testing.B, apps, systemEntities int, perApp bool) {
	ids := make([]string, systemEntities)
	for i := range ids {
		ids[i] = "cmd-" + string(rune('a'+i%26)) + string(rune('a'+(i/26)%26)) + string(rune('a'+(i/676)%26))
	}

	registry := make([]*App, apps)
	for i := range registry {
		app := NewApp("app", Env{})
		app.AddCommand(testCommand(ids[i%len(ids)], "name"))
		registry[i] = app
	}

	b.ResetTimer()
	for b.Loop() {
		if perApp {
			// Old shape: every app rebuilds the system-wide set.
			for _, app := range registry {
				app.RemoveDanglingCommands(idLookupSet(ids))
			}
		} else {
			// Current shape: built once, shared by every app.
			set := idLookupSet(ids)
			for _, app := range registry {
				app.RemoveDanglingCommands(set)
			}
		}
	}
}

func BenchmarkSweepSetPerApp(b *testing.B) { benchmarkSweep(b, 500, 5000, true) }
func BenchmarkSweepSetShared(b *testing.B) { benchmarkSweep(b, 500, 5000, false) }
