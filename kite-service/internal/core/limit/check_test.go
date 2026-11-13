package limit

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name   string
		counts store.ItemCounts
		added  store.ItemCounts
		ok     bool
	}{
		{"under limit", store.ItemCounts{Total: 14, Enabled: 14}, store.ItemCounts{Total: 1, Enabled: 1}, true},
		{"at limit", store.ItemCounts{Total: 15, Enabled: 15}, store.ItemCounts{Total: 1, Enabled: 1}, false},
		{"disabled don't count", store.ItemCounts{Total: 30, Enabled: 14}, store.ItemCounts{Total: 1, Enabled: 1}, true},
		{"add disabled at limit", store.ItemCounts{Total: 15, Enabled: 15}, store.ItemCounts{Total: 1}, true},
		{"add disabled over limit", store.ItemCounts{Total: 20, Enabled: 20}, store.ItemCounts{Total: 1}, true},
		{"enable at limit", store.ItemCounts{Total: 20, Enabled: 15}, store.ItemCounts{Enabled: 1}, false},
		{"enable under limit at total", store.ItemCounts{Total: 45, Enabled: 14}, store.ItemCounts{Enabled: 1}, true},
		{"total reached", store.ItemCounts{Total: 45}, store.ItemCounts{Total: 1}, false},
		{"import over limit", store.ItemCounts{Total: 10, Enabled: 10}, store.ItemCounts{Total: 6, Enabled: 6}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Check("commands", 15, tt.counts, tt.added)
			if tt.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
