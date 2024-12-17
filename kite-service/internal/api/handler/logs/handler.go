package logs

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type LogHandler struct {
	logStore store.LogStore
}

func NewLogHandler(logStore store.LogStore) *LogHandler {
	return &LogHandler{
		logStore: logStore,
	}
}

func (h *LogHandler) HandleLogSummaryGet(c *handler.Context) (*wire.LogSummaryGetResponse, error) {
	entries, err := h.logStore.LogSummary(c.Context(), c.App.ID, time.Now().UTC().Add(-time.Hour*24), time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to get log entries: %w", err)
	}

	return wire.LogSummaryToWire(entries), nil
}

func (h *LogHandler) HandleLogEntryList(c *handler.Context) (*wire.LogEntryListResponse, error) {
	beforeID, _ := strconv.ParseInt(c.Query("before"), 10, 64)

	entries, err := h.logStore.LogEntriesByApp(c.Context(), c.App.ID, beforeID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get log entries: %w", err)
	}

	res := make([]*wire.LogEntry, len(entries))
	for i, entry := range entries {
		res[i] = wire.LogEntryToWire(entry)
	}

	return &res, nil
}
