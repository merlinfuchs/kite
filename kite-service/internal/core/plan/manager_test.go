package plan

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

func testManager() *PlanManager {
	return &PlanManager{
		plans: []model.Plan{
			{ID: "basic", Default: true, FeatureUsageCreditsPerMonth: 10000, FeatureMaxGuilds: 100},
			{ID: "pro", FeatureUsageCreditsPerMonth: 100000, FeatureMaxGuilds: 500},
		},
	}
}

// The usage sweep skips the entitlement lookup for any app under this floor,
// which is only sound because Merge takes the maximum of each field and the
// default plan is always merged in.
func TestDefaultFeaturesIsTheFloor(t *testing.T) {
	m := testManager()

	base := m.DefaultFeatures()
	if base.UsageCreditsPerMonth != 10000 {
		t.Errorf("default credits = %d, want the default plan's 10000", base.UsageCreditsPerMonth)
	}

	entitled := m.featuresFromEntitlements([]*model.Entitlement{{PlanID: "pro"}})
	if entitled.UsageCreditsPerMonth < base.UsageCreditsPerMonth {
		t.Errorf("entitled credits %d fell below the default floor %d",
			entitled.UsageCreditsPerMonth, base.UsageCreditsPerMonth)
	}
	if entitled.UsageCreditsPerMonth != 100000 {
		t.Errorf("entitled credits = %d, want the pro plan's 100000", entitled.UsageCreditsPerMonth)
	}
}

// An entitlement naming a plan that no longer exists must not strip an app
// back below the default.
func TestFeaturesFromUnknownEntitlementKeepsDefault(t *testing.T) {
	m := testManager()

	got := m.featuresFromEntitlements([]*model.Entitlement{{PlanID: "removed-plan"}})
	if got != m.DefaultFeatures() {
		t.Errorf("unknown entitlement produced %+v, want the default features", got)
	}
}

// With no default plan configured the floor is zero, so the sweep's filter
// degenerates to checking every app rather than skipping any.
func TestDefaultFeaturesWithNoDefaultPlan(t *testing.T) {
	m := &PlanManager{plans: []model.Plan{{ID: "pro", FeatureUsageCreditsPerMonth: 100000}}}

	if got := m.DefaultFeatures().UsageCreditsPerMonth; got != 0 {
		t.Errorf("floor with no default plan = %d, want 0", got)
	}
}
