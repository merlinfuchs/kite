package engine

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	arikawastate "github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
	"gopkg.in/guregu/null.v4"
)

type fakeTimerStore struct {
	store.ResumePointStore

	mu     sync.Mutex
	timers []*model.ResumePoint
}

func (f *fakeTimerStore) CreateResumePoint(ctx context.Context, resumePoint *model.ResumePoint) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.timers = append(f.timers, resumePoint)
	return nil
}

func (f *fakeTimerStore) CountPendingTimerResumePoints(ctx context.Context, appID string) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.timers), nil
}

func (f *fakeTimerStore) HasDueTimerResumePoints(ctx context.Context, now time.Time) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, rp := range f.timers {
		if !rp.ResumeAt.Time.After(now) {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeTimerStore) LeaseDueTimerResumePoints(ctx context.Context, appIDs []string, now time.Time, leaseUntil time.Time, batchSize int) ([]*model.ResumePoint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var due []*model.ResumePoint
	for _, rp := range f.timers {
		if !rp.ResumeAt.Time.After(now) {
			rp.ResumeAt = null.TimeFrom(leaseUntil)
			leased := *rp
			due = append(due, &leased)
		}
	}
	return due, nil
}

func (f *fakeTimerStore) DeleteTimerResumePoint(ctx context.Context, appID string, id string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, rp := range f.timers {
		if rp.ID == id {
			f.timers = append(f.timers[:i], f.timers[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeTimerStore) pending() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.timers)
}

func TestScheduledFlowResumesAfterDurableSleep(t *testing.T) {
	listener := scheduledListener("* * * * *")
	listener.FlowSource = flow.FlowData{
		Nodes: []flow.FlowNode{
			{ID: "entry", Type: flow.FlowNodeTypeEntryEvent, Data: flow.FlowNodeData{
				EventType:         flow.EventTypeScheduleCron,
				EventScheduleCron: "* * * * *",
				Description:       "test",
			}},
			{ID: "sleep", Type: flow.FlowNodeTypeControlSleep, Data: flow.FlowNodeData{SleepDurationSeconds: "3600"}},
			{ID: "log", Type: flow.FlowNodeTypeActionLog, Data: flow.FlowNodeData{
				LogLevel:   provider.LogLevelInfo,
				LogMessage: "woke up from {{schedule.time}}",
			}},
		},
		Edges: []flow.FlowEdge{
			{Source: "entry", Target: "sleep"},
			{Source: "sleep", Target: "log"},
		},
	}

	listeners := &fakeScheduleListenerStore{
		lastRuns:  make(map[string]time.Time),
		listeners: []*model.EventListener{listener},
	}
	logs := &fakeLogStore{entries: make(chan model.LogEntry, 1)}
	timers := &fakeTimerStore{}

	e := newScheduleTestEngine(listeners, logs)
	e.env.ResumePointStore = timers
	e.populate(context.Background())

	s := newTestScheduler(e, fakeSessions{})

	ctx := context.Background()
	s.tick(ctx, at(12, 0, 30))
	s.tick(ctx, at(12, 1, 0))

	// Wait for the scheduled run to suspend.
	deadline := time.Now().Add(5 * time.Second)
	for timers.pending() != 1 {
		if time.Now().After(deadline) {
			t.Fatal("scheduled flow didn't suspend")
		}
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case entry := <-logs.entries:
		t.Fatalf("blocks after the sleep ran early: %q", entry.Message)
	default:
	}

	// Not due yet.
	s.resumeTimers(ctx, time.Now().UTC())
	if timers.pending() != 1 {
		t.Fatal("timer was resumed before it was due")
	}

	s.resumeTimers(ctx, time.Now().UTC().Add(time.Hour+time.Minute))

	select {
	case entry := <-logs.entries:
		// The resumed flow still sees the schedule event it was started with.
		if entry.Message != "woke up from 2026-09-25T12:01:00Z" {
			t.Fatalf("log message = %q", entry.Message)
		}
		if entry.EventListenerID.String != "listener" {
			t.Fatalf("log not attributed to the listener: %+v", entry)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("flow didn't resume after the sleep")
	}

	deadline = time.Now().Add(5 * time.Second)
	for timers.pending() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("resumed timer wasn't deleted")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestTimerIsRetriedWhenItCantResume(t *testing.T) {
	logs := &fakeLogStore{entries: make(chan model.LogEntry, 1)}
	timers := &fakeTimerStore{timers: []*model.ResumePoint{{
		ID:              "timer",
		Type:            model.ResumePointTypeTimer,
		AppID:           "app",
		EventListenerID: null.StringFrom("listener"),
		FlowNodeID:      "sleep",
		ResumeAt:        null.TimeFrom(at(12, 0, 0)),
	}}}

	// The engine hasn't loaded the listener yet, e.g. right after a restart.
	e := newScheduleTestEngine(&fakeScheduleListenerStore{lastRuns: make(map[string]time.Time)}, logs)
	e.env.ResumePointStore = timers
	s := newTestScheduler(e, fakeSessions{})

	s.resumeTimers(context.Background(), at(12, 0, 1))
	time.Sleep(100 * time.Millisecond)

	if timers.pending() != 1 {
		t.Fatal("timer that couldn't resume was deleted")
	}
	if got := timers.timers[0].ResumeAt.Time; !got.Equal(at(12, 0, 1).Add(timerLease)) {
		t.Fatalf("timer resumes at %s, want after the lease", got)
	}
}

func TestTimerResumeRestoresInteractionToken(t *testing.T) {
	crypt, err := util.NewSymmetricCrypt(strings.Repeat("ab", 32))
	if err != nil {
		t.Fatal(err)
	}
	timers := &fakeTimerStore{}
	env := Env{ResumePointStore: timers, TokenCrypt: crypt}

	state := flow.NewFlowContextState()
	state.Triggers = []flow.FlowTrigger{{Interaction: &discord.InteractionEvent{ID: 1}}}
	// Stored triggers never hold tokens.
	state.ResumeTrigger = &flow.FlowTrigger{Interaction: &discord.InteractionEvent{ID: 2}}

	p := NewResumePointProvider(timers, crypt, "app", entityLinks{})
	_, err = p.CreateResumePoint(context.Background(), flow.ResumePoint{
		Type:             flow.ResumePointTypeTimer,
		NodeID:           "sleep",
		State:            *state,
		ResumeAt:         time.Now().Add(time.Hour),
		InteractionToken: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}

	rp := timers.timers[0]
	if rp.InteractionToken.String == "secret" {
		t.Fatal("interaction token stored in plain text")
	}

	event, restored, err := NewApp("app", env).timerResumeEvent(rp)
	if err != nil {
		t.Fatal(err)
	}

	e, ok := event.(*gateway.InteractionCreateEvent)
	if !ok {
		t.Fatalf("event is %T", event)
	}
	if e.ID != 2 || e.Token != "secret" {
		t.Fatalf("restored interaction %d with token %q, want 2 with the stored token", e.ID, e.Token)
	}
	if len(restored.Triggers) != 1 || restored.Triggers[0].Interaction.ID != 1 {
		t.Fatalf("earlier triggers = %+v, want the first unchanged", restored.Triggers)
	}
	if restored.ResumeTrigger != nil {
		t.Fatal("resume trigger is still set in the resumed state")
	}
}

func TestTimerLimitPerApp(t *testing.T) {
	timers := &fakeTimerStore{}
	for range maxPendingTimers {
		timers.timers = append(timers.timers, &model.ResumePoint{})
	}

	p := NewResumePointProvider(timers, nil, "app", entityLinks{})
	_, err := p.CreateResumePoint(context.Background(), flow.ResumePoint{
		Type:     flow.ResumePointTypeTimer,
		ResumeAt: time.Now().Add(time.Hour),
	})
	if err == nil {
		t.Fatal("created a timer beyond the limit")
	}
}

func TestTimerOfDeletedListenerIsDropped(t *testing.T) {
	timers := &fakeTimerStore{timers: []*model.ResumePoint{{
		ID:         "timer",
		Type:       model.ResumePointTypeTimer,
		AppID:      "app",
		FlowNodeID: "sleep",
		// The listener was deleted, which cleared the link.
		ResumeAt: null.TimeFrom(at(12, 0, 0)),
	}}}

	e := newScheduleTestEngine(&fakeScheduleListenerStore{lastRuns: make(map[string]time.Time)}, &fakeLogStore{})
	e.env.ResumePointStore = timers
	s := newTestScheduler(e, fakeSessions{})

	s.resumeTimers(context.Background(), at(12, 0, 1))

	deadline := time.Now().Add(5 * time.Second)
	for timers.pending() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("timer of a deleted listener wasn't dropped")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (f *fakeTimerStore) ResumePoint(ctx context.Context, appID string, id string) (*model.ResumePoint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, rp := range f.timers {
		if rp.ID == id {
			copied := *rp
			return &copied, nil
		}
	}
	return nil, store.ErrNotFound
}

func (f *fakeTimerStore) TouchResumePoint(ctx context.Context, appID string, id string, usedAt time.Time) error {
	return nil
}

func TestClickResumeCountsDurableSleepsFromZero(t *testing.T) {
	listener := scheduledListener("* * * * *")
	// Keeps the ID of the log block it replaces, so the edges still apply.
	listener.FlowSource.Nodes[1] = flow.FlowNode{
		ID:   "log",
		Type: flow.FlowNodeTypeControlSleep,
		Data: flow.FlowNodeData{SleepDurationSeconds: "3600"},
	}

	state := flow.NewFlowContextState()
	state.DurableSleeps = 5

	// A menu button's resume point, saved after 5 durable sleeps.
	timers := &fakeTimerStore{timers: []*model.ResumePoint{{
		ID:              "button",
		Type:            model.ResumePointTypeMessageComponents,
		AppID:           "app",
		EventListenerID: null.StringFrom("listener"),
		FlowNodeID:      "log",
		FlowState:       *state,
	}}}

	e := newScheduleTestEngine(&fakeScheduleListenerStore{
		lastRuns:  make(map[string]time.Time),
		listeners: []*model.EventListener{listener},
	}, &fakeLogStore{})
	e.env.ResumePointStore = timers
	e.populate(context.Background())

	app := e.appForID("app")
	if !app.resumeFlow("button", arikawastate.New("Bot test"), &schedule.Event{Time: at(12, 0, 0)}) {
		t.Fatal("resume point wasn't found")
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		timers.mu.Lock()
		var timer *model.ResumePoint
		for _, rp := range timers.timers {
			if rp.ID != "button" {
				timer = rp
			}
		}
		timers.mu.Unlock()

		if timer != nil {
			if got := timer.FlowState.DurableSleeps; got != 1 {
				t.Fatalf("durable sleeps after the click = %d, want 1", got)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("clicked branch didn't sleep")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
