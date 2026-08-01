// Package metrics exposes lightweight runtime counters for the gateway and
// engine. Everything here is published through expvar and served by the debug
// server (see internal/entry/server), which is disabled by default and should
// only ever be bound to a private interface.
//
// The counters are deliberately cheap: they sit on the per-event hot path, so
// they must not allocate or lock. expvar.Map.Add is a sync.Map read plus an
// atomic add for keys that already exist, which is acceptable at event rates.
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

	// EngineDispatchCount and EngineDispatchNanos together give the mean
	// dispatch duration. expvar has no histogram; mean is enough to compare
	// before and after a change.
	EngineDispatchCount = expvar.NewInt("engine_dispatch_count")
	EngineDispatchNanos = expvar.NewInt("engine_dispatch_nanos_total")

	// EngineLockWaitCount and EngineLockWaitNanos measure contention on the
	// engine's registry lock. These replace the ad-hoc slog.Warn blocks that
	// only fired past a fixed threshold.
	EngineLockWaitCount = expvar.NewInt("engine_lock_wait_count")
	EngineLockWaitNanos = expvar.NewInt("engine_lock_wait_nanos_total")

	// DBPollCount and DBPollNanos are keyed by poll name (for example
	// "populate_commands") so the polling loops can be compared before and
	// after their intervals change.
	DBPollCount = expvar.NewMap("db_poll_count")
	DBPollNanos = expvar.NewMap("db_poll_nanos_total")
)

func init() {
	expvar.Publish("go_goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
}

// ObserveDispatch records the duration of a single engine event dispatch.
func ObserveDispatch(start time.Time) {
	EngineDispatchCount.Add(1)
	EngineDispatchNanos.Add(int64(time.Since(start)))
}

// ObserveLockWait records how long a caller waited to acquire the engine lock.
func ObserveLockWait(start time.Time) {
	EngineLockWaitCount.Add(1)
	EngineLockWaitNanos.Add(int64(time.Since(start)))
}

// ObservePoll records the duration of a database polling query.
func ObservePoll(name string, start time.Time) {
	DBPollCount.Add(name, 1)
	DBPollNanos.Add(name, int64(time.Since(start)))
}
