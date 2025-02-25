package billing

import (
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func (h *BillingHandler) HandleBillingPlanList(c *handler.Context) (*wire.BillingPlanListResponse, error) {
	plans := h.planManager.Plans()

	res := make(wire.BillingPlanListResponse, len(plans))
	for i, plan := range plans {
		p := wire.BillingPlan(plan)
		res[i] = &p
	}

	return &res, nil
}
