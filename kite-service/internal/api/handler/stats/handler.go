package stats

import (
	"fmt"
	"sync"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// The counts scan large tables, and the website doesn't need them to the minute.
const cacheDuration = time.Hour

type StatsHandler struct {
	statsStore store.StatsStore

	mu       sync.Mutex
	cached   *wire.StatsGetResponse
	cachedAt time.Time
}

func NewStatsHandler(statsStore store.StatsStore) *StatsHandler {
	return &StatsHandler{statsStore: statsStore}
}

func (h *StatsHandler) HandleStatsGet(c *handler.Context) (*wire.StatsGetResponse, error) {
	// Holding the lock while querying keeps concurrent requests from all
	// running the query when the cache expires.
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cached != nil && time.Since(h.cachedAt) < cacheDuration {
		return h.cached, nil
	}

	stats, err := h.statsStore.Stats(c.Context(), time.Now().UTC().AddDate(0, 0, -30))
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	h.cached = &wire.StatsGetResponse{
		Users:               stats.Users,
		Apps:                stats.Apps,
		Commands:            stats.Commands,
		EventListeners:      stats.EventListeners,
		ExecutionsLastMonth: stats.Executions,
	}
	h.cachedAt = time.Now()
	return h.cached, nil
}
