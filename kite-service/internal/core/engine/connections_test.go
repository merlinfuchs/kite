package engine

import (
	"testing"
	"time"
)

func TestConnectionTracker(t *testing.T) {
	tracker := NewConnectionTracker()

	if !tracker.ConnectedAt("app").IsZero() {
		t.Error("an app that never connected should have no connection time")
	}

	before := time.Now()
	tracker.Connected("app")

	if got := tracker.ConnectedAt("app"); got.Before(before) || got.After(time.Now()) {
		t.Errorf("connection time: got %s", got)
	}
	if !tracker.ConnectedAt("other").IsZero() {
		t.Error("connecting one app shouldn't affect another")
	}

	var none *ConnectionTracker
	if !none.ConnectedAt("app").IsZero() {
		t.Error("a nil tracker should know of no connection")
	}
}
