// Package metrics exposes lightweight runtime counters for the gateway and
// engine. Everything here is published through expvar and served by the debug
// server (see internal/entry/server), which is disabled by default and should
// only ever be bound to a private interface.
//
// Only GatewayEvents sits on the per-event path, and it is deliberately the
// only one: expvar counters are process-global 8-byte values, so every core
// touching one invalidates the same cache line. A handful of them per event
// measurably outweighs the work they were measuring. Lock contention is
// reported by the existing threshold logging instead.
package metrics

import (
	"expvar"
	"runtime"
	"time"
)

var (
	// GatewayEvents counts every event received from Discord, keyed by event
	// type. This is the baseline for measuring intent narrowing: after
	// per-app intents ship, MESSAGE_CREATE here should fall sharply.
	GatewayEvents = expvar.NewMap("gateway_events_total")

	// GatewayEventsDropped counts events that were received but not
	// dispatched, keyed by reason. A non-trivial "unknown_app" count means
	// the gateway connected before the engine loaded that app's entities.
	GatewayEventsDropped = expvar.NewMap("gateway_events_dropped_total")

	// GatewayConnections tracks the number of live gateway connections owned
	// by this process.
	GatewayConnections = expvar.NewInt("gateway_connections_active")

	// GatewayIntentReconnects counts reconnects caused by an app's intent
	// requirements changing. A sustained rate here means something is
	// flip-flopping and reconnecting apps in a loop.
	GatewayIntentReconnects = expvar.NewInt("gateway_intent_reconnects_total")

	// DBPollCount and DBPollNanos are keyed by poll name (for example
	// "populate_commands") so the polling loops can be compared before and
	// after their intervals change.
	DBPollCount = expvar.NewMap("db_poll_count")
	DBPollNanos = expvar.NewMap("db_poll_nanos_total")

	// SlowLockWaits counts lock acquisitions that took longer than
	// SlowLockThreshold, keyed by lock name.
	//
	// Counting only slow acquisitions rather than timing every one keeps this
	// off the fast path: the duration is already measured for the threshold
	// logging, and the counter is only touched when something is wrong. That
	// is also why the threshold here can be far more sensitive than the one
	// guarding the log lines — a counter cannot flood anything.
	SlowLockWaits = expvar.NewMap("lock_wait_slow_total")
)

// SlowLockThreshold is the point past which a lock acquisition is counted as
// slow. These locks only guard in-memory map access, where a normal
// acquisition is tens of nanoseconds, so a millisecond is already four orders
// of magnitude beyond normal while staying clear of ordinary goroutine
// preemption jitter.
const SlowLockThreshold = time.Millisecond

func init() {
	expvar.Publish("go_goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
}

// ObservePoll records the duration of a database polling query.
func ObservePoll(name string, start time.Time) {
	DBPollCount.Add(name, 1)
	DBPollNanos.Add(name, int64(time.Since(start)))
}

// ObserveLockWait counts the acquisition if it was slow. Callers pass the
// duration they already measured rather than a start time, so no additional
// clock read lands on the hot path.
func ObserveLockWait(name string, waited time.Duration) {
	if waited > SlowLockThreshold {
		SlowLockWaits.Add(name, 1)
	}
}
