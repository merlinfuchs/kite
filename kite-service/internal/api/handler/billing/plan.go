package billing

import (
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func (h *BillingHandler) HandleBillingPlanList(c *handler.Context) (*wire.BillingPlanListResponse, error) {
	res := make(wire.BillingPlanListResponse, len(h.config.Plans))
	for i, plan := range h.config.Plans {
		p := wire.BillingPlan(plan)
		res[i] = &p
	}

	return &res, nil
}
