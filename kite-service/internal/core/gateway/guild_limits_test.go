package gateway

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type fakeEntitlementStore struct {
	store.EntitlementStore
	entitlements map[string][]*model.Entitlement
	err          error
	calls        int
}

func (s *fakeEntitlementStore) ActiveEntitlementsForApps(ctx context.Context, appIDs []string, now time.Time) (map[string][]*model.Entitlement, error) {
	s.calls++
	return s.entitlements, s.err
}

type fakeDisableAppStore struct {
	store.AppStore
	mu       sync.Mutex
	disabled []string
}

func (s *fakeDisableAppStore) DisableApp(ctx context.Context, opts store.AppDisableOpts) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disabled = append(s.disabled, opts.ID)
	return true, nil
}

type fakeLogStore struct {
	store.LogStore
}

func (fakeLogStore) CreateLogEntry(ctx context.Context, entry model.LogEntry) error {
	return nil
}

func newGuildLimitManager(entitlements *fakeEntitlementStore, apps *fakeDisableAppStore, guilds map[string]int) *GatewayManager {
	plans := []model.Plan{
		{ID: "free", Default: true, FeatureMaxGuilds: 10},
		{ID: "paid", FeatureMaxGuilds: 100},
	}

	m := &GatewayManager{
		planManager: plan.NewPlanManager(entitlements, nil, nil, nil, plans, plan.PlanManagerConfig{}),
		gateways:    map[string]*Gateway{},
	}
	for appID, count := range guilds {
		g := &Gateway{appID: appID, appStore: apps, logStore: fakeLogStore{}}
		g.readyGuilds.Store(int64(count))
		m.gateways[appID] = g
	}
	return m
}

func TestEnforceGuildLimitsDisablesAppsOverTheirPlan(t *testing.T) {
	entitlements := &fakeEntitlementStore{entitlements: map[string][]*model.Entitlement{
		"paid_under": {{AppID: "paid_under", PlanID: "paid"}},
		"paid_over":  {{AppID: "paid_over", PlanID: "paid"}},
	}}
	apps := &fakeDisableAppStore{}
	m := newGuildLimitManager(entitlements, apps, map[string]int{
		"free_under": 5,
		"free_over":  20,
		"paid_under": 50,
		"paid_over":  200,
	})

	if err := m.enforceGuildLimits(context.Background()); err != nil {
		t.Fatal(err)
	}

	disabled := map[string]bool{}
	for _, id := range apps.disabled {
		disabled[id] = true
	}
	want := map[string]bool{"free_over": true, "paid_over": true}
	if len(disabled) != len(want) || !disabled["free_over"] || !disabled["paid_over"] {
		t.Errorf("disabled %v, want %v", apps.disabled, want)
	}

	// Disabled apps aren't disabled again before their gateway is stopped.
	apps.disabled = nil
	if err := m.enforceGuildLimits(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(apps.disabled) != 0 {
		t.Errorf("disabled again: %v", apps.disabled)
	}
}

// When the plans can't be looked up, paid apps must not be held to the
// default plan, like during a database outage while every gateway reconnects.
func TestEnforceGuildLimitsSkipsWhenPlansFail(t *testing.T) {
	entitlements := &fakeEntitlementStore{err: errors.New("database unavailable")}
	apps := &fakeDisableAppStore{}
	m := newGuildLimitManager(entitlements, apps, map[string]int{"paid": 50})

	if err := m.enforceGuildLimits(context.Background()); err == nil {
		t.Error("expected an error")
	}
	if len(apps.disabled) != 0 {
		t.Errorf("disabled %v", apps.disabled)
	}
}

func TestEnforceGuildLimitsSkipsLookupUnderDefaultLimit(t *testing.T) {
	entitlements := &fakeEntitlementStore{}
	m := newGuildLimitManager(entitlements, &fakeDisableAppStore{}, map[string]int{"a": 3, "b": 10})

	if err := m.enforceGuildLimits(context.Background()); err != nil {
		t.Fatal(err)
	}
	if entitlements.calls != 0 {
		t.Errorf("looked up plans %d times", entitlements.calls)
	}
}
