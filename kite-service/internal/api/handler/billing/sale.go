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
