package billing

import (
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func TestSaleActive(t *testing.T) {
	start := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	sale := config.BillingSaleConfig{Name: "Black Week", Code: "BW", Percent: 30, StartsAt: start, EndsAt: end}
	withDiscount := sale
	withDiscount.PaddleDiscountID = "dsc_1"

	tests := []struct {
		name     string
		provider model.SubscriptionSource
		sale     config.BillingSaleConfig
		now      time.Time
		want     bool
	}{
		{"during the sale", "", sale, start, true},
		{"before the sale", "", sale, start.Add(-time.Second), false},
		{"after the sale", "", sale, end, false},
		{"on Lemon Squeezy", model.SubscriptionSourceLemonSqueezy, sale, start, true},
		{"on Paddle without a discount", model.SubscriptionSourcePaddle, sale, start, false},
		{"on Paddle with a discount", model.SubscriptionSourcePaddle, withDiscount, start, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewBillingHandler(BillingHandlerConfig{Sale: tt.sale, CheckoutProvider: tt.provider}, nil, nil, nil, nil, nil, nil)
			h.now = func() time.Time { return tt.now }
			if got := h.saleActive(); got != tt.want {
				t.Errorf("saleActive() = %v, want %v", got, tt.want)
			}
		})
	}
}
