package billing

import (
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func TestSaleDiscountCode(t *testing.T) {
	plans := []model.Plan{{
		ID:                          "plus",
		LemonSqueezyVariantID:       "10",
		LemonSqueezyYearlyVariantID: "11",
	}}
	planManager := plan.NewPlanManager(nil, nil, nil, nil, plans, plan.PlanManagerConfig{})
	start := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	h := NewBillingHandler(BillingHandlerConfig{Sale: config.BillingSaleConfig{
		Name: "Black Week", Code: "BW", Percent: 30, StartsAt: start, EndsAt: end,
	}}, nil, nil, planManager, nil, nil)

	tests := []struct {
		name    string
		now     time.Time
		variant string
		want    string
	}{
		{"yearly during the sale", start, "11", "BW"},
		{"monthly during the sale", start, "10", ""},
		{"unknown variant", start, "12", ""},
		{"yearly before the sale", start.Add(-time.Second), "11", ""},
		{"yearly after the sale", end, "11", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.now = func() time.Time { return tt.now }
			if got := h.saleDiscountCode(tt.variant); got != tt.want {
				t.Errorf("saleDiscountCode(%q) = %q, want %q", tt.variant, got, tt.want)
			}
		})
	}
}
