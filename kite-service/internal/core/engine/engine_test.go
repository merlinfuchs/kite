package engine

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// The fakes embed their store interface so any method the engine does not call
// panics rather than silently returning a zero value.

type fakeCommandStore struct {
	store.CommandStore
	commands   []*model.Command
	err        error
	calledWith []time.Time
}

func (f *fakeCommandStore) CommandsUpdatedSince(ctx context.Context, since time.Time) ([]*model.Command, error) {
	f.calledWith = append(f.calledWith, since)
	return f.commands, f.err
}

type fakeEventListenerStore struct {
	store.EventListenerStore
	err error
}

func (f *fakeEventListenerStore) EventListenersUpdatedSince(ctx context.Context, since time.Time) ([]*model.EventListener, error) {
	return nil, f.err
}

type fakePluginInstanceStore struct {
	store.PluginInstanceStore
	err error
}

func (f *fakePluginInstanceStore) PluginInstancesUpdatedSince(ctx context.Context, since time.Time) ([]*model.PluginInstance, error) {
	return nil, f.err
}

type fakeDeletedEntityStore struct {
	store.DeletedEntityStore
	deleted []*model.DeletedEntity
	err     error
}

func (f *fakeDeletedEntityStore) DeletedEntitiesSince(ctx context.Context, since time.Time) ([]*model.DeletedEntity, error) {
	return f.deleted, f.err
}

func newTestEngine(commands *fakeCommandStore, listeners *fakeEventListenerStore, plugins *fakePluginInstanceStore) *Engine {
	return newTestEngineWithDeleted(commands, listeners, plugins, &fakeDeletedEntityStore{})
}

func newTestEngineWithDeleted(
	commands *fakeCommandStore,
	listeners *fakeEventListenerStore,
	plugins *fakePluginInstanceStore,
	deleted *fakeDeletedEntityStore,
) *Engine {
	return NewEngine(Env{
		Config: EngineConfig{
			ClusterCount:    1,
			ClusterIndex:    0,
			PopulateOverlap: 5 * time.Second,
		},
		CommandStore:        commands,
		EventListenerStore:  listeners,
		PluginInstanceStore: plugins,
		DeletedEntityStore:  deleted,
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

// The deleted entity query counts like the others: a failure leaves its window
// unread.
func TestPopulateDeletedEntitiesFailureDoesNotAdvanceCursor(t *testing.T) {
	e := newTestEngineWithDeleted(
		&fakeCommandStore{},
		&fakeEventListenerStore{},
		&fakePluginInstanceStore{},
		&fakeDeletedEntityStore{err: errors.New("connection refused")},
	)
	cursor := time.Now().UTC().Add(-time.Minute)
	e.lastUpdate = cursor

	e.populate(context.Background())

	if !e.lastUpdate.Equal(cursor) {
		t.Errorf("cursor moved to %v despite the deleted entity query failing", e.lastUpdate)
	}
}

func TestPopulateDropsDeletedEntities(t *testing.T) {
	deleted := &fakeDeletedEntityStore{}
	e := newTestEngineWithDeleted(&fakeCommandStore{}, &fakeEventListenerStore{}, &fakePluginInstanceStore{}, deleted)

	app := e.appForID("app")
	app.AddCommand(testCommand("cmd", "ping"))
	app.AddEventListener(testListener("listener", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))
	e.lastUpdate = time.Now().UTC().Add(-time.Minute)

	deleted.deleted = []*model.DeletedEntity{
		{ID: "cmd", Type: model.DeletedEntityTypeCommand, AppID: "app"},
		{ID: "listener", Type: model.DeletedEntityTypeEventListener, AppID: "app"},
	}
	e.populate(context.Background())

	if got := app.commandsByName["ping"]; got != nil {
		t.Error("deleted command is still loaded")
	}
	if got := app.listenersByType[model.EventListenerTypeDiscordMessageCreate]; len(got) != 0 {
		t.Error("deleted event listener is still loaded")
	}
}

func TestPopulateDropsDisabledCommand(t *testing.T) {
	commands := &fakeCommandStore{}
	e := newTestEngine(commands, &fakeEventListenerStore{}, &fakePluginInstanceStore{})

	app := e.appForID("app")
	app.AddCommand(testCommand("cmd", "ping"))

	commands.commands = []*model.Command{{ID: "cmd", AppID: "app", Name: "ping", Enabled: false}}
	e.populate(context.Background())

	if got := app.commandsByName["ping"]; got != nil {
		t.Error("disabled command is still loaded")
	}
}

// Reproduces a dangling sweep at production shape: many apps in the registry,
// each tested against the system-wide set of enabled entities. Measures the
// steady state where nothing is dangling, which is the common case.
//
// The set is built once per sweep. Building it per app -- the previous
// behaviour -- made this O(apps x entities) and accounted for 74% of the
// process's CPU in production.
func BenchmarkDanglingSweep(b *testing.B) {
	const (
		apps           = 500
		systemEntities = 5000
	)

	enabledIDs := make([]string, systemEntities)
	for i := range enabledIDs {
		enabledIDs[i] = "cmd-" + strconv.Itoa(i)
	}

	registry := make([]*App, apps)
	for i := range registry {
		app := NewApp("app", Env{})
		app.AddCommand(testCommand(enabledIDs[i%len(enabledIDs)], "name"))
		registry[i] = app
	}

	for b.Loop() {
		set := util.IDSet(enabledIDs)
		for _, app := range registry {
			app.RemoveDanglingCommands(set)
		}
	}
}

// Interactions for message template buttons and resume points are resolved from
// the database, so an app without commands, listeners or plugins still needs
// to receive them.
func TestHandleEventRegistersAppForInteraction(t *testing.T) {
	e := newTestEngine(&fakeCommandStore{}, &fakeEventListenerStore{}, &fakePluginInstanceStore{})

	e.HandleEvent("app", nil, &gateway.InteractionCreateEvent{
		InteractionEvent: discord.InteractionEvent{ID: discord.InteractionID(discord.NewSnowflake(time.Now()))},
	})

	if _, ok := e.apps["app"]; !ok {
		t.Fatal("expected the app to be registered for an interaction")
	}
}

func TestHandleEventDropsOtherEventsForUnknownApp(t *testing.T) {
	e := newTestEngine(&fakeCommandStore{}, &fakeEventListenerStore{}, &fakePluginInstanceStore{})

	e.HandleEvent("app", nil, &gateway.MessageCreateEvent{})

	if _, ok := e.apps["app"]; ok {
		t.Fatal("expected no app to be registered for a non-interaction event")
	}
}
