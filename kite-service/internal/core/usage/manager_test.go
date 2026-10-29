package usage

import (
	"context"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type fakeAppStore struct {
	store.AppStore
	// paused holds the apps paused for credits.
	paused map[string]bool
	// warned holds the apps whose 80% DM was sent this month.
	warned map[string]bool
	// noticesOff holds the apps whose owner turned the DMs off.
	noticesOff map[string]bool
}

func (s *fakeAppStore) App(ctx context.Context, id string) (*model.App, error) {
	return &model.App{ID: id, CreditNoticesEnabled: !s.noticesOff[id]}, nil
}

func (s *fakeAppStore) ClaimAppCreditsWarning(ctx context.Context, appID string, monthStart time.Time, now time.Time) (bool, error) {
	if s.warned[appID] {
		return false, nil
	}
	s.warned[appID] = true
	return true, nil
}

func (s *fakeAppStore) PauseAppForCredits(ctx context.Context, appID string, now time.Time) (bool, error) {
	if s.paused[appID] {
		return false, nil
	}
	s.paused[appID] = true
	return true, nil
}

func (s *fakeAppStore) UnpauseAppForCredits(ctx context.Context, appID string, now time.Time) (bool, error) {
	if !s.paused[appID] {
		return false, nil
	}
	delete(s.paused, appID)
	return true, nil
}

func (s *fakeAppStore) CreditsPausedAppIDs(ctx context.Context) ([]string, error) {
	return slices.Collect(maps.Keys(s.paused)), nil
}

type fakeUsageStore struct {
	store.UsageStore
	used map[string]int
}

func (s *fakeUsageStore) AllUsageCreditsUsedBetween(ctx context.Context, start time.Time, end time.Time) (map[string]int, error) {
	return s.used, nil
}

type fakeLogStore struct {
	store.LogStore
	entries []model.LogEntry
}

func (s *fakeLogStore) CreateLogEntry(ctx context.Context, entry model.LogEntry) error {
	s.entries = append(s.entries, entry)
	return nil
}

type fakeCreditNotifier struct {
	low    []string
	usedUp []string
}

func (n *fakeCreditNotifier) CreditsLow(app *model.App, used int, limit int, resetAt time.Time) {
	n.low = append(n.low, app.ID)
}

func (n *fakeCreditNotifier) CreditsUsedUp(app *model.App, limit int, resetAt time.Time) {
	n.usedUp = append(n.usedUp, app.ID)
}

type fakeEntitlementStore struct {
	store.EntitlementStore
	plans map[string]string
}

func (s *fakeEntitlementStore) ActiveEntitlementsForApps(ctx context.Context, appIDs []string, now time.Time) (map[string][]*model.Entitlement, error) {
	res := map[string][]*model.Entitlement{}
	for _, id := range appIDs {
		if p, ok := s.plans[id]; ok {
			res[id] = []*model.Entitlement{{AppID: id, PlanID: p}}
		}
	}
	return res, nil
}

func TestUpdateAppsForCredits(t *testing.T) {
	apps := &fakeAppStore{paused: map[string]bool{
		"paused_new_month": true,
		"paused_upgraded":  true,
		"paused_still_out": true,
	}, warned: map[string]bool{
		"warned_before": true,
	}, noticesOff: map[string]bool{
		"over_notices_off": true,
	}}
	usage := &fakeUsageStore{used: map[string]int{
		"under":            500,
		"over":             10000,
		"pro_over_free":    20000,
		"paused_upgraded":  15000,
		"paused_still_out": 12000,
		"near":             8000,
		"warned_before":    9000,
		"pro_near_free":    9000,
		"over_notices_off": 10000,
	}}
	logs := &fakeLogStore{}
	entitlements := &fakeEntitlementStore{plans: map[string]string{
		"pro_over_free":   "pro",
		"paused_upgraded": "pro",
		"pro_near_free":   "pro",
	}}

	planManager := plan.NewPlanManager(entitlements, nil, nil, nil, []model.Plan{
		{ID: "basic", Default: true, FeatureUsageCreditsPerMonth: 10000},
		{ID: "pro", FeatureUsageCreditsPerMonth: 100000},
	}, plan.PlanManagerConfig{})

	notifier := &fakeCreditNotifier{}
	m := NewUsageManager(apps, usage, logs, nil, nil, nil, nil, nil, nil, nil, planManager, notifier)

	if err := m.updateAppsForCredits(context.Background()); err != nil {
		t.Fatal(err)
	}

	paused := slices.Sorted(maps.Keys(apps.paused))
	want := []string{"over", "over_notices_off", "paused_still_out"}
	if !slices.Equal(paused, want) {
		t.Errorf("paused apps = %v, want %v", paused, want)
	}

	var logged []string
	for _, e := range logs.entries {
		logged = append(logged, e.AppID)
	}
	slices.Sort(logged)
	want = []string{"over", "over_notices_off", "paused_new_month", "paused_upgraded"}
	if !slices.Equal(logged, want) {
		t.Errorf("logged apps = %v, want %v", logged, want)
	}

	if want := []string{"near"}; !slices.Equal(notifier.low, want) {
		t.Errorf("warned apps = %v, want %v", notifier.low, want)
	}
	if want := []string{"over"}; !slices.Equal(notifier.usedUp, want) {
		t.Errorf("paused apps notified = %v, want %v", notifier.usedUp, want)
	}

	// Nothing changed, so a second sweep logs nothing again.
	logs.entries = nil
	if err := m.updateAppsForCredits(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(logs.entries) != 0 {
		t.Errorf("second sweep logged %d entries, want none", len(logs.entries))
	}
	if len(notifier.low) != 1 || len(notifier.usedUp) != 1 {
		t.Errorf("second sweep sent DMs: low %v, used up %v", notifier.low, notifier.usedUp)
	}
}
