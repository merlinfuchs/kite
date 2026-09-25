package engine

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
	"gopkg.in/guregu/null.v4"
)

const (
	// scheduleGracePeriod is how far back a run missed during downtime is
	// still caught up on. Only the latest missed run is caught up on.
	scheduleGracePeriod = 5 * time.Minute

	scheduleTickInterval     = time.Second
	scheduleFlushInterval    = 5 * time.Second
	scheduleFeaturesCacheTTL = time.Minute
)

type SessionProvider interface {
	AppSession(ctx context.Context, appID string) (*state.State, error)
}

// listenerSchedule is the run state of a scheduled event listener. It's only
// touched by the scheduler loop, except when a reloaded listener takes over
// the state of the one it replaces.
type listenerSchedule struct {
	sync.Mutex

	cron  string
	sched *schedule.Schedule

	// lastRun is the occurrence that last ran, not when it ran.
	lastRun time.Time
	// next is the first occurrence that hasn't run yet. Zero until the
	// scheduler first sees the listener.
	next time.Time
	// catchUp allows running the latest occurrence missed within the grace
	// period. Only set for listeners loaded at startup, so enabling or editing
	// a listener doesn't immediately run an occurrence from before the change.
	catchUp bool

	// running is shared with the schedules that replace this one on reload,
	// so a run that's still going also blocks the reloaded listener.
	running *atomic.Bool
	// retired is set once a reloaded listener took over, so a tick that still
	// holds this one can't run an occurrence the new one runs too.
	retired bool
}

func newListenerSchedule(cron string, lastRun null.Time, catchUp bool) (*listenerSchedule, error) {
	sched, err := schedule.Parse(cron)
	if err != nil {
		return nil, err
	}

	return &listenerSchedule{
		cron:    cron,
		sched:   sched,
		lastRun: lastRun.Time,
		catchUp: catchUp && lastRun.Valid,
		running: new(atomic.Bool),
	}, nil
}

// takeOver continues from the state of the schedule this one replaces, so a
// reload doesn't lose runs that haven't been persisted yet.
func (s *listenerSchedule) takeOver(old *listenerSchedule) {
	old.Lock()
	lastRun, next := old.lastRun, old.next
	old.retired = true
	old.Unlock()

	s.Lock()
	defer s.Unlock()

	if lastRun.After(s.lastRun) {
		s.lastRun = lastRun
	}
	// Keep the pending occurrence, otherwise a reload right before it would
	// skip it. It only applies if the schedule itself didn't change.
	if old.cron == s.cron {
		s.next = next
	}
	s.catchUp = false
	s.running = old.running
}

// start sets the first occurrence when the scheduler first sees the listener.
// It's separate from due so it also happens while the app's gateway is down.
func (s *listenerSchedule) start(now time.Time) {
	s.Lock()
	defer s.Unlock()
	s.startLocked(now)
}

func (s *listenerSchedule) startLocked(now time.Time) {
	if !s.next.IsZero() {
		return
	}

	since := now
	if s.catchUp {
		since = now.Add(-scheduleGracePeriod)
		if s.lastRun.After(since) {
			since = s.lastRun
		}
	}
	s.next = s.sched.Next(since)
}

// due returns the occurrence to run at now, if there is one. Occurrences that
// were missed are skipped except for the latest one, and only if it's within
// the grace period. Runs closer than minInterval to the previous one are
// skipped too, which covers apps whose plan changed after the schedule was
// saved. A returned occurrence counts as run.
func (s *listenerSchedule) due(now time.Time, minInterval time.Duration) (time.Time, bool) {
	s.Lock()
	defer s.Unlock()

	s.startLocked(now)
	if s.retired || s.next.IsZero() || now.Before(s.next) {
		return time.Time{}, false
	}

	// After a long gateway outage, stepping through every missed occurrence
	// of a frequent schedule would stall the scheduler. Anything older than
	// the grace period is skipped anyway.
	if now.Sub(s.next) > scheduleGracePeriod {
		s.next = s.sched.Next(now.Add(-scheduleGracePeriod))
		if s.next.IsZero() || now.Before(s.next) {
			return time.Time{}, false
		}
	}

	// Skip to the latest occurrence that's due.
	occurrence := s.next
	s.next = s.sched.Next(occurrence)
	for !s.next.IsZero() && !s.next.After(now) {
		occurrence, s.next = s.next, s.sched.Next(s.next)
	}

	if now.Sub(occurrence) > scheduleGracePeriod {
		return time.Time{}, false
	}
	if !s.lastRun.IsZero() && occurrence.Sub(s.lastRun) < minInterval {
		return time.Time{}, false
	}

	s.lastRun = occurrence
	return occurrence, true
}

// RunScheduler runs scheduled event listeners until ctx is done. Sessions come
// from the gateway manager, which is created after the engine.
func (e *Engine) RunScheduler(ctx context.Context, sessions SessionProvider) {
	s := &scheduler{
		engine:   e,
		sessions: sessions,
		features: e.env.FeatureProvider,
		lastRuns: make(map[string]time.Time),
	}
	go s.run(ctx)
	// Database writes run separately, so a slow database doesn't delay runs.
	go s.runFlush(ctx)
}

type scheduler struct {
	engine   *Engine
	sessions SessionProvider
	features FeatureProvider

	// Features of all apps with scheduled listeners, refreshed in one batch in
	// the background so per-second schedules don't query entitlements every
	// second and a slow query doesn't delay runs.
	featuresMu       sync.Mutex
	cachedFeatures   map[string]model.Features
	featuresCachedAt time.Time
	refreshing       atomic.Bool

	// lastRuns are persisted in batches, per-second schedules would otherwise
	// write to the database every second.
	lastRunsMu sync.Mutex
	lastRuns   map[string]time.Time
}

func (s *scheduler) run(ctx context.Context) {
	ticker := time.NewTicker(scheduleTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.tick(ctx, now.UTC())
		}
	}
}

func (s *scheduler) runFlush(ctx context.Context) {
	ticker := time.NewTicker(scheduleFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// ctx is already done, so the final flush needs its own.
			s.flush(context.Background())
			return
		case <-ticker.C:
			s.flush(ctx)
		}
	}
}

func (s *scheduler) tick(ctx context.Context, now time.Time) {
	listeners := s.engine.scheduledEventListeners()
	if len(listeners) == 0 {
		return
	}

	features := s.appFeatures(ctx, listeners, now)
	sessions := make(map[string]*state.State)

	for _, l := range listeners {
		appID := l.listener.AppID

		// The session is checked before due so the occurrence stays due until
		// the gateway is up, e.g. while gateways are still starting after a
		// restart. due drops it once it's older than the grace period.
		l.schedule.start(now)
		session, ok := sessions[appID]
		if !ok {
			session, _ = s.sessions.AppSession(ctx, appID)
			sessions[appID] = session
		}
		if session == nil {
			continue
		}

		occurrence, ok := l.schedule.due(now, features[appID].MinScheduleInterval())
		if !ok {
			continue
		}
		s.lastRunsMu.Lock()
		s.lastRuns[l.listener.ID] = occurrence
		s.lastRunsMu.Unlock()

		// A slow flow on a frequent schedule would otherwise pile up runs.
		if !l.schedule.running.CompareAndSwap(false, true) {
			continue
		}

		go func() {
			defer l.schedule.running.Store(false)
			l.HandleScheduledRun(session, occurrence)
		}()
	}
}

// appFeatures returns the cached features and starts a refresh in the
// background when they're stale or miss an app. Apps without cached features
// get zero features, which means the strictest interval.
func (s *scheduler) appFeatures(ctx context.Context, listeners []*EventListener, now time.Time) map[string]model.Features {
	s.featuresMu.Lock()
	cached := s.cachedFeatures
	stale := now.Sub(s.featuresCachedAt) >= scheduleFeaturesCacheTTL
	s.featuresMu.Unlock()

	if !stale {
		for _, l := range listeners {
			if _, ok := cached[l.listener.AppID]; !ok {
				stale = true
				break
			}
		}
	}

	if stale && s.refreshing.CompareAndSwap(false, true) {
		appIDSet := make(map[string]struct{}, len(listeners))
		for _, l := range listeners {
			appIDSet[l.listener.AppID] = struct{}{}
		}
		go s.refreshFeatures(ctx, slices.Collect(maps.Keys(appIDSet)), now)
	}

	return cached
}

func (s *scheduler) refreshFeatures(ctx context.Context, appIDs []string, now time.Time) {
	defer s.refreshing.Store(false)

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	fetched, err := s.features.AppFeaturesForApps(ctx, appIDs)

	s.featuresMu.Lock()
	defer s.featuresMu.Unlock()

	// A failed refresh keeps the old features until the next refresh is due,
	// instead of retrying every tick.
	s.featuresCachedAt = now

	if err != nil {
		slog.Error(
			"Failed to get app features for scheduled event listeners",
			slog.String("error", err.Error()),
		)

		// Apps without cached features get zero features, so they don't make
		// the cache look stale on every tick until the next refresh.
		features := maps.Clone(s.cachedFeatures)
		if features == nil {
			features = make(map[string]model.Features, len(appIDs))
		}
		for _, appID := range appIDs {
			if _, ok := features[appID]; !ok {
				features[appID] = model.Features{}
			}
		}
		s.cachedFeatures = features
		return
	}

	s.cachedFeatures = fetched
}

func (s *scheduler) flush(ctx context.Context) {
	s.lastRunsMu.Lock()
	lastRuns := s.lastRuns
	s.lastRuns = make(map[string]time.Time)
	s.lastRunsMu.Unlock()

	if len(lastRuns) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.engine.env.EventListenerStore.UpdateEventListenersLastRunAt(ctx, lastRuns)
	if err != nil {
		slog.Error(
			"Failed to persist last runs of scheduled event listeners",
			slog.String("error", err.Error()),
		)

		// Kept for the next flush, unless the listener ran again meanwhile. At
		// worst a run missed during downtime is caught up on twice.
		s.lastRunsMu.Lock()
		for id, t := range lastRuns {
			if _, ok := s.lastRuns[id]; !ok {
				s.lastRuns[id] = t
			}
		}
		s.lastRunsMu.Unlock()
	}
}
