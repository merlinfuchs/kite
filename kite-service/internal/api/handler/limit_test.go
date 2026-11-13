package handler

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestCheckEnabledLimit(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		counts       store.ItemCounts
		added        int
		addedEnabled int
		ok           bool
	}{
		{"no limit", 0, store.ItemCounts{Total: 100, Enabled: 100}, 1, 1, true},
		{"under limit", 15, store.ItemCounts{Total: 14, Enabled: 14}, 1, 1, true},
		{"at limit", 15, store.ItemCounts{Total: 15, Enabled: 15}, 1, 1, false},
		{"disabled don't count", 15, store.ItemCounts{Total: 30, Enabled: 14}, 1, 1, true},
		{"add disabled at limit", 15, store.ItemCounts{Total: 15, Enabled: 15}, 1, 0, true},
		{"add disabled over limit", 15, store.ItemCounts{Total: 20, Enabled: 20}, 1, 0, true},
		{"enable at limit", 15, store.ItemCounts{Total: 20, Enabled: 15}, 0, 1, false},
		{"enable under limit at total", 15, store.ItemCounts{Total: 45, Enabled: 14}, 0, 1, true},
		{"total reached", 15, store.ItemCounts{Total: 45, Enabled: 0}, 1, 0, false},
		{"import over limit", 15, store.ItemCounts{Total: 10, Enabled: 10}, 6, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckEnabledLimit("commands", tt.limit, tt.counts, tt.added, tt.addedEnabled)
			if tt.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
