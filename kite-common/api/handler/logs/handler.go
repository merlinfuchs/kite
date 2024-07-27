package logs

import (
	"fmt"
	"strconv"

	"github.com/kitecloud/kite/kite-common/api/handler"
	"github.com/kitecloud/kite/kite-common/api/wire"
	"github.com/kitecloud/kite/kite-common/store"
)

type LogHandler struct {
	logStore store.LogStore
}

func NewLogHandler(logStore store.LogStore) *LogHandler {
	return &LogHandler{
		logStore: logStore,
	}
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
