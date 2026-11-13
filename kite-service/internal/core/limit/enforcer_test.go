package limit

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAppStore struct {
	store.AppStore
	overSince map[string]time.Time
}

func (s *fakeAppStore) App(ctx context.Context, id string) (*model.App, error) {
	return &model.App{ID: id}, nil
}

func (s *fakeAppStore) MarkAppOverLimit(ctx context.Context, appID string, now time.Time) (bool, error) {
	if _, ok := s.overSince[appID]; ok {
		return false, nil
	}
	s.overSince[appID] = now
	return true, nil
}

func (s *fakeAppStore) AppsOverLimitSince(ctx context.Context, appIDs []string) (map[string]time.Time, error) {
	res := map[string]time.Time{}
	for _, id := range appIDs {
		if t, ok := s.overSince[id]; ok {
			res[id] = t
		}
	}
	return res, nil
}

func (s *fakeAppStore) ClearAppsOverLimit(ctx context.Context, keepAppIDs []string) error {
	keep := map[string]bool{}
	for _, id := range keepAppIDs {
		keep[id] = true
	}
	for id := range s.overSince {
		if !keep[id] {
			delete(s.overSince, id)
		}
	}
	return nil
}

type fakeCommandStore struct {
	store.CommandStore
	enabled map[string]int
}

func (s *fakeCommandStore) AppsWithEnabledCommandsOver(ctx context.Context, minCount int) (map[string]int, error) {
	res := map[string]int{}
	for id, n := range s.enabled {
		if n > minCount {
			res[id] = n
		}
	}
	return res, nil
}

func (s *fakeCommandStore) DisableNewestCommands(ctx context.Context, appID string, count int, updatedAt time.Time) (int, error) {
	n := min(count, s.enabled[appID])
	s.enabled[appID] -= n
	return n, nil
}

type fakeEventListenerStore struct {
	store.EventListenerStore
	enabled map[string]map[model.EventSource]int
}

func (s *fakeEventListenerStore) AppsWithEnabledEventListenersOver(ctx context.Context, minCount int) ([]store.EnabledEventListenerCount, error) {
	var res []store.EnabledEventListenerCount
	for id, bySource := range s.enabled {
		for source, n := range bySource {
			if n > minCount {
				res = append(res, store.EnabledEventListenerCount{AppID: id, Source: source, Count: n})
			}
		}
	}
	return res, nil
}

func (s *fakeEventListenerStore) DisableNewestEventListeners(ctx context.Context, appID string, source model.EventSource, count int, updatedAt time.Time) (int, error) {
	n := min(count, s.enabled[appID][source])
	s.enabled[appID][source] -= n
	return n, nil
}

type fakeLogStore struct {
	store.LogStore
	entries []model.LogEntry
}

func (s *fakeLogStore) CreateLogEntry(ctx context.Context, entry model.LogEntry) error {
	s.entries = append(s.entries, entry)
	return nil
}

type fakeFeatures struct {
	paid map[string]bool
}

var (
	freeFeatures = model.Features{MaxCommands: 15, MaxEventListeners: 10, MaxScheduledEventListeners: 5, MaxWebhookEventListeners: 5}
	paidFeatures = model.Features{MaxCommands: 50, MaxEventListeners: 15, MaxScheduledEventListeners: 5, MaxWebhookEventListeners: 5}
)

func (f *fakeFeatures) DefaultFeatures() model.Features {
	return freeFeatures
}

func (f *fakeFeatures) AppFeaturesForApps(ctx context.Context, appIDs []string) (map[string]model.Features, error) {
	res := map[string]model.Features{}
	for _, id := range appIDs {
		res[id] = freeFeatures
		if f.paid[id] {
			res[id] = paidFeatures
		}
	}
	return res, nil
}

type notice struct {
	appID    string
	enforced bool
	details  string
}

type fakeNotifier struct {
	notices []notice
}

func (n *fakeNotifier) OverLimit(app *model.App, details string) {
	n.notices = append(n.notices, notice{appID: app.ID, details: details})
}

func (n *fakeNotifier) LimitEnforced(app *model.App, details string) {
	n.notices = append(n.notices, notice{appID: app.ID, enforced: true, details: details})
}

type env struct {
	apps      *fakeAppStore
	commands  *fakeCommandStore
	listeners *fakeEventListenerStore
	logs      *fakeLogStore
	features  *fakeFeatures
	notifier  *fakeNotifier
	enforcer  *Enforcer
}

func newEnv() *env {
	e := &env{
		apps:      &fakeAppStore{overSince: map[string]time.Time{}},
		commands:  &fakeCommandStore{enabled: map[string]int{}},
		listeners: &fakeEventListenerStore{enabled: map[string]map[model.EventSource]int{}},
		logs:      &fakeLogStore{},
		features:  &fakeFeatures{paid: map[string]bool{}},
		notifier:  &fakeNotifier{},
	}
	e.enforcer = NewEnforcer(e.apps, e.commands, e.listeners, e.logs, e.features, e.notifier)
	return e
}

func TestEnforceWarnsThenDisablesNewest(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	start := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)

	e.commands.enabled["app"] = 18
	e.listeners.enabled["app"] = map[model.EventSource]int{model.EventSourceDiscord: 12, model.EventSourceSchedule: 2}

	require.NoError(t, e.enforcer.Enforce(ctx, start))
	require.Len(t, e.notifier.notices, 1)
	assert.False(t, e.notifier.notices[0].enforced)
	assert.Equal(t,
		"has 18 enabled commands and 12 enabled event listeners, but its plan allows 15 commands and 10 event listeners. Disable or delete some, or upgrade its plan. Otherwise Kite disables the newest ones on November 16 (UTC).",
		e.notifier.notices[0].details,
	)
	require.Len(t, e.logs.entries, 1)
	assert.Equal(t, "Your app "+e.notifier.notices[0].details, e.logs.entries[0].Message)

	// Warned once, nothing disabled during the grace period.
	require.NoError(t, e.enforcer.Enforce(ctx, start.Add(GracePeriod-time.Hour)))
	assert.Len(t, e.notifier.notices, 1)
	assert.Equal(t, 18, e.commands.enabled["app"])

	require.NoError(t, e.enforcer.Enforce(ctx, start.Add(GracePeriod)))
	require.Len(t, e.notifier.notices, 2)
	assert.True(t, e.notifier.notices[1].enforced)
	assert.Equal(t,
		"had more enabled commands and event listeners than its plan allows, so Kite disabled the 3 newest commands and the 2 newest event listeners. You can enable them again after you upgrade its plan or disable others.",
		e.notifier.notices[1].details,
	)
	assert.Equal(t, 15, e.commands.enabled["app"])
	assert.Equal(t, 10, e.listeners.enabled["app"][model.EventSourceDiscord])
	assert.Equal(t, 2, e.listeners.enabled["app"][model.EventSourceSchedule])

	// Under the limits now, so the mark is cleared and nothing else happens.
	require.NoError(t, e.enforcer.Enforce(ctx, start.Add(GracePeriod+time.Hour)))
	assert.Len(t, e.notifier.notices, 2)
	assert.Empty(t, e.apps.overSince)
}

func TestEnforceClearsAfterUpgrade(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	start := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)

	e.commands.enabled["app"] = 16
	require.NoError(t, e.enforcer.Enforce(ctx, start))
	assert.Contains(t, e.apps.overSince, "app")

	e.features.paid["app"] = true
	require.NoError(t, e.enforcer.Enforce(ctx, start.Add(GracePeriod)))
	assert.Equal(t, 16, e.commands.enabled["app"])
	assert.NotContains(t, e.apps.overSince, "app")
	assert.Len(t, e.notifier.notices, 1)

	// Going over again starts a new grace period.
	e.features.paid["app"] = false
	require.NoError(t, e.enforcer.Enforce(ctx, start.Add(GracePeriod+time.Hour)))
	assert.Len(t, e.notifier.notices, 2)
	assert.Equal(t, 16, e.commands.enabled["app"])
}

func TestEnforceIgnoresAppsWithinLimits(t *testing.T) {
	e := newEnv()
	ctx := context.Background()

	e.commands.enabled["paid"] = 40
	e.features.paid["paid"] = true
	e.commands.enabled["free"] = 15
	e.listeners.enabled["free"] = map[model.EventSource]int{model.EventSourceWebhook: 5}

	require.NoError(t, e.enforcer.Enforce(ctx, time.Now()))
	assert.Empty(t, e.notifier.notices)
	assert.Empty(t, e.apps.overSince)
}

func TestDetailsOfOneKind(t *testing.T) {
	assert.Equal(t,
		"had more enabled webhooks than its plan allows, so Kite disabled the newest one. You can enable it again after you upgrade its plan or disable others.",
		enforcedDetails([]exceeded{{source: model.EventSourceWebhook, count: 1, limit: 5}}),
	)
	assert.Equal(t,
		"has 18 enabled commands, but its plan allows 15. Disable or delete some, or upgrade its plan. Otherwise Kite disables the newest ones on November 16 (UTC).",
		overLimitDetails([]exceeded{{count: 18, limit: 15}}, time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)),
	)
}
