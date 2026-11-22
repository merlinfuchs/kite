package billing

import (
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func (h *BillingHandler) HandleBillingSaleGet(c *handler.Context) (*wire.BillingSaleGetResponse, error) {
	sale := h.config.Sale
	if !sale.Active(h.now()) {
		return &wire.BillingSaleGetResponse{}, nil
	}

	return &wire.BillingSaleGetResponse{
		Sale: &wire.BillingSale{
			Name:    sale.Name,
			Code:    sale.Code,
			Percent: sale.Percent,
			EndsAt:  sale.EndsAt,
		},
	}, nil
}

// saleDiscountCode is the code a checkout of the variant applies, empty when
// no sale runs or the variant isn't yearly.
func (h *BillingHandler) saleDiscountCode(variantID string) string {
	if !h.config.Sale.Active(h.now()) {
		return ""
	}
	plan := h.planManager.PlanByLemonSqueezyVariantID(variantID)
	if plan == nil || plan.LemonSqueezyYearlyVariantID != variantID {
		return ""
	}
	return h.config.Sale.Code
}
