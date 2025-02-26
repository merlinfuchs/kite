package billing

import (
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

func (h *BillingHandler) HandleFeaturesGet(c *handler.Context) (*wire.FeaturesGetResponse, error) {
	features := h.planManager.AppFeatures(c.Context(), c.App.ID)

	res := wire.Features(features)
	return &res, nil
}
