package billing

import (
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func (h *BillingHandler) HandleFeaturesGet(c *handler.Context) (*wire.FeaturesGetResponse, error) {
	res := wire.Features(c.Features)
	return &res, nil
}
