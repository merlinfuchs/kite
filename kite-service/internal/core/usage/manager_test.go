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
	// reasons holds the disabled reason of each stopped app, apps not in it
	// are enabled.
	reasons map[string]string
}

func (s *fakeAppStore) DisableApp(ctx context.Context, opts store.AppDisableOpts) (bool, error) {
	if _, ok := s.reasons[opts.ID]; ok {
		return false, nil
	}
	s.reasons[opts.ID] = opts.DisabledReason.String
	return true, nil
}

func (s *fakeAppStore) EnableAppDisabledWithReason(ctx context.Context, appID string, reason string, updatedAt time.Time) (bool, error) {
	if r, ok := s.reasons[appID]; !ok || r != reason {
		return false, nil
	}
	delete(s.reasons, appID)
	return true, nil
}

func (s *fakeAppStore) AppIDsDisabledWithReason(ctx context.Context, reason string) ([]string, error) {
	var ids []string
	for id, r := range s.reasons {
		if r == reason {
			ids = append(ids, id)
		}
	}
	return ids, nil
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
	apps := &fakeAppStore{reasons: map[string]string{
		"stopped_new_month": model.AppDisabledReasonNoCredits,
		"stopped_upgraded":  model.AppDisabledReasonNoCredits,
		"stopped_still_out": model.AppDisabledReasonNoCredits,
		"stopped_by_user":   "",
		"stopped_token":     model.AppDisabledReasonInvalidToken,
	}}
	usage := &fakeUsageStore{used: map[string]int{
		"under":             500,
		"over":              10000,
		"pro_over_free":     20000,
		"stopped_upgraded":  15000,
		"stopped_still_out": 12000,
	}}
	logs := &fakeLogStore{}
	entitlements := &fakeEntitlementStore{plans: map[string]string{
		"pro_over_free":    "pro",
		"stopped_upgraded": "pro",
	}}

	planManager := plan.NewPlanManager(entitlements, nil, nil, []model.Plan{
		{ID: "basic", Default: true, FeatureUsageCreditsPerMonth: 10000},
		{ID: "pro", FeatureUsageCreditsPerMonth: 100000},
	}, plan.PlanManagerConfig{})

	m := NewUsageManager(apps, usage, logs, nil, nil, nil, planManager)

	if err := m.updateAppsForCredits(context.Background()); err != nil {
		t.Fatal(err)
	}

	stopped := slices.Sorted(maps.Keys(apps.reasons))
	want := []string{"over", "stopped_by_user", "stopped_still_out", "stopped_token"}
	if !slices.Equal(stopped, want) {
		t.Errorf("stopped apps = %v, want %v", stopped, want)
	}

	var logged []string
	for _, e := range logs.entries {
		logged = append(logged, e.AppID)
	}
	slices.Sort(logged)
	want = []string{"over", "stopped_new_month", "stopped_upgraded"}
	if !slices.Equal(logged, want) {
		t.Errorf("logged apps = %v, want %v", logged, want)
	}

	// Nothing changed, so a second sweep logs nothing again.
	logs.entries = nil
	if err := m.updateAppsForCredits(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(logs.entries) != 0 {
		t.Errorf("second sweep logged %d entries, want none", len(logs.entries))
	}
}
