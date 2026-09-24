package metrics

import (
	"testing"
	"time"
)

// The counter must stay silent for normal acquisitions, or it stops being a
// signal. These locks guard in-memory maps, so the common case is nanoseconds.
func TestObserveLockWaitIgnoresFastAcquisitions(t *testing.T) {
	const name = "test_fast"

	ObserveLockWait(name, 50*time.Nanosecond)
	ObserveLockWait(name, SlowLockThreshold)

	if got := SlowLockWaits.Get(name); got != nil {
		t.Errorf("fast acquisitions were counted: %v", got)
	}
}

func TestObserveLockWaitCountsSlowAcquisitions(t *testing.T) {
	const name = "test_slow"

	ObserveLockWait(name, SlowLockThreshold+time.Millisecond)
	ObserveLockWait(name, time.Second)

	got := SlowLockWaits.Get(name)
	if got == nil {
		t.Fatal("slow acquisitions were not counted")
	}
	if got.String() != "2" {
		t.Errorf("counted %s slow acquisitions, want 2", got)
	}
}

// A threshold at or below ordinary goroutine preemption jitter would count
// noise rather than contention.
func TestSlowLockThresholdIsAboveSchedulerJitter(t *testing.T) {
	if SlowLockThreshold < 100*time.Microsecond {
		t.Errorf("SlowLockThreshold = %v, low enough to count preemption jitter", SlowLockThreshold)
	}
}
