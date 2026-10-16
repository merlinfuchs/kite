package engine

import (
	"sync"
	"time"
)

// ConnectionTracker remembers when each app's bot last connected to Discord,
// which the connection itself doesn't keep.
type ConnectionTracker struct {
	mu          sync.RWMutex
	connectedAt map[string]time.Time
}

func NewConnectionTracker() *ConnectionTracker {
	return &ConnectionTracker{
		connectedAt: make(map[string]time.Time),
	}
}

// Connected records that the app's bot connected now. A nil tracker records
// nothing.
func (t *ConnectionTracker) Connected(appID string) {
	if t == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.connectedAt[appID] = time.Now()
}

// ConnectedAt returns when the app's bot last connected, or the zero time if
// it hasn't since the service started. A nil tracker knows of no connection.
func (t *ConnectionTracker) ConnectedAt(appID string) time.Time {
	if t == nil {
		return time.Time{}
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.connectedAt[appID]
}
