package billing

import (
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func (h *BillingHandler) HandleBillingSaleGet(c *handler.Context) (*wire.BillingSaleGetResponse, error) {
	if !h.saleActive() {
		return &wire.BillingSaleGetResponse{}, nil
	}

	sale := h.config.Sale
	return &wire.BillingSaleGetResponse{
		Sale: &wire.BillingSale{
			Name:    sale.Name,
			Code:    sale.Code,
			Percent: sale.Percent,
			EndsAt:  sale.EndsAt,
		},
	}, nil
}

// saleActive is whether a sale runs now that checkouts of yearly plans can
// apply. Paddle checkouts need the sale's Paddle discount.
func (h *BillingHandler) saleActive() bool {
	if !h.config.Sale.Active(h.now()) {
		return false
	}
	return h.checkoutProvider() != model.SubscriptionSourcePaddle || h.config.Sale.PaddleDiscountID != ""
}
