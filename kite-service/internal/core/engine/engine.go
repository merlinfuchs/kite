package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/metrics"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type Engine struct {
	sync.RWMutex

	env Env

	lastUpdate time.Time
	apps       map[string]*App

	// scheduledApps are the apps that had a scheduled listener added, so the
	// scheduler doesn't have to walk every app each second. Apps whose
	// scheduled listeners are all gone are pruned by the scheduler.
	scheduledAppsMu sync.Mutex
	scheduledApps   map[string]*App
}

func NewEngine(
	env Env,
) *Engine {
	if env.BlockRateLimiter == nil {
		env.BlockRateLimiter = NewBlockRateLimiter()
	}

	return &Engine{
		env:           env,
		apps:          make(map[string]*App),
		scheduledApps: make(map[string]*App),
	}
}

func (e *Engine) Run(ctx context.Context) {
	populateInterval := util.IntervalOrDefault(e.env.Config.PopulateInterval, 5*time.Second)
	removeDanglingInterval := util.IntervalOrDefault(e.env.Config.RemoveDanglingInterval, 10*time.Minute)

	go func() {
		updateTicker := time.NewTicker(populateInterval)
		defer updateTicker.Stop()

		removeTicker := time.NewTicker(removeDanglingInterval)
		defer removeTicker.Stop()

		// Populate immediately rather than waiting out the first tick, so
		// events arriving right after startup have somewhere to go. Anything
		// dispatched before this completes is counted as unknown_app.
		e.populate(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-updateTicker.C:
				e.populate(ctx)
			case <-removeTicker.C:
				e.removeDangling(ctx)
			}
		}
	}()
}

// populate loads entities changed since the last successful poll.
//
// The cursor is captured before the queries run and only advanced once they
// all succeed. Advancing it beforehand (or after a failure) silently drops
// every entity written while the queries were in flight, because the next poll
// would already consider that window covered.
func (e *Engine) populate(ctx context.Context) {
	tickStart := time.Now().UTC()
	lastUpdate := e.lastUpdate

	ok := true

	if err := e.populatePlugins(ctx, lastUpdate); err != nil {
		ok = false
		slog.Error(
			"Failed to populate plugins in engine",
			slog.String("error", err.Error()),
		)
	}
	if err := e.populateCommands(ctx, lastUpdate); err != nil {
		ok = false
		slog.Error(
			"Failed to populate commands in engine",
			slog.String("error", err.Error()),
		)
	}
	if err := e.populateEventListeners(ctx, lastUpdate); err != nil {
		ok = false
		slog.Error(
			"Failed to populate event listeners in engine",
			slog.String("error", err.Error()),
		)
	}
	// Must run after the queries above, so an entity that was changed and then
	// deleted within the window ends up removed.
	if err := e.removeDeletedEntities(ctx, lastUpdate); err != nil {
		ok = false
		slog.Error(
			"Failed to remove deleted entities in engine",
			slog.String("error", err.Error()),
		)
	}

	if !ok {
		// Leave the cursor where it is so the next poll retries this window.
		return
	}

	// Rewind by the overlap so rows committed out of timestamp order, or
	// during the query itself, are still picked up. Populating is idempotent,
	// so re-reading a few rows costs nothing.
	e.lastUpdate = tickStart.Add(-e.env.Config.PopulateOverlap)
}

// removeDangling is a backstop for entities whose disable or delete the
// polls missed.
func (e *Engine) removeDangling(ctx context.Context) {
	e.env.BlockRateLimiter.Sweep()

	if err := e.removeDanglingPlugins(ctx); err != nil {
		slog.Error(
			"Failed to remove dangling plugins in engine",
			slog.String("error", err.Error()),
		)
	}
	if err := e.removeDanglingCommands(ctx); err != nil {
		slog.Error(
			"Failed to remove dangling commands in engine",
			slog.String("error", err.Error()),
		)
	}
	if err := e.removeDanglingEventListeners(ctx); err != nil {
		slog.Error(
			"Failed to remove dangling event listeners in engine",
			slog.String("error", err.Error()),
		)
	}
}

// existingApp returns the app with the given ID, or nil if the engine has
// nothing loaded for it.
func (e *Engine) existingApp(appID string) *App {
	e.RLock()
	defer e.RUnlock()
	return e.apps[appID]
}

// appForID returns the app with the given ID, creating it if necessary. The
// registry lock is held only for the map access, never across flow
// compilation or plugin construction.
func (e *Engine) appForID(appID string) *App {
	// The app almost always exists; a miss only happens on its first entity.
	// Taking the write lock unconditionally would block every concurrent
	// dispatch, since Go's RWMutex gives waiting writers priority over readers.
	e.RLock()
	app, ok := e.apps[appID]
	e.RUnlock()

	if ok {
		return app
	}

	lockStart := time.Now()
	e.Lock()
	defer e.Unlock()
	metrics.ObserveLockWait("engine_registry_write", time.Since(lockStart))

	// Re-check: another goroutine may have created it while the lock was free.
	if app, ok := e.apps[appID]; ok {
		return app
	}

	app = NewApp(appID, e.env)
	e.apps[appID] = app

	return app
}

func (e *Engine) populatePlugins(ctx context.Context, lastUpdate time.Time) error {
	queryStart := time.Now()
	pluginInstances, err := e.env.PluginInstanceStore.PluginInstancesUpdatedSince(ctx, lastUpdate)
	metrics.ObservePoll("populate_plugins", queryStart)
	if err != nil {
		return fmt.Errorf("failed to get plugin instances: %w", err)
	}

	var disabled []*model.DeletedEntity
	for _, pluginInstance := range pluginInstances {
		if util.CluserForKey(pluginInstance.AppID, e.env.Config.ClusterCount) != e.env.Config.ClusterIndex {
			continue
		}

		if !pluginInstance.Enabled {
			disabled = append(disabled, &model.DeletedEntity{
				ID:    pluginInstance.ID,
				Type:  model.DeletedEntityTypePluginInstance,
				AppID: pluginInstance.AppID,
			})
			continue
		}

		// AddPluginInstance may construct a plugin instance, which the plugin
		// interface permits to do I/O. It takes the app's own lock, never the
		// registry lock.
		e.appForID(pluginInstance.AppID).AddPluginInstance(pluginInstance)
	}

	e.removeEntities(disabled)
	return nil
}

func (e *Engine) removeDanglingPlugins(ctx context.Context) error {
	pluginInstanceIDs, err := e.env.PluginInstanceStore.EnabledPluginInstanceIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled plugin instance IDs: %w", err)
	}

	idSet := util.IDSet(pluginInstanceIDs)

	e.RLock()
	defer e.RUnlock()

	for _, app := range e.apps {
		app.RemoveDanglingPluginInstances(idSet)
	}

	return nil
}

func (e *Engine) populateCommands(ctx context.Context, lastUpdate time.Time) error {
	queryStart := time.Now()
	commands, err := e.env.CommandStore.CommandsUpdatedSince(ctx, lastUpdate)
	metrics.ObservePoll("populate_commands", queryStart)
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	var disabled []*model.DeletedEntity
	for _, command := range commands {
		if util.CluserForKey(command.AppID, e.env.Config.ClusterCount) != e.env.Config.ClusterIndex {
			continue
		}

		if !command.Enabled {
			disabled = append(disabled, &model.DeletedEntity{
				ID:    command.ID,
				Type:  model.DeletedEntityTypeCommand,
				AppID: command.AppID,
			})
			continue
		}

		// Compile before touching the registry. Flow compilation used to run
		// under the engine's write lock, blocking every concurrent dispatch.
		compiled, err := NewCommand(command, e.env)
		if err != nil {
			// NewCommand already logged the compilation failure.
			continue
		}

		e.appForID(command.AppID).AddCommand(command.ID, compiled)
	}

	e.removeEntities(disabled)
	return nil
}

func (e *Engine) removeDanglingCommands(ctx context.Context) error {
	commandIDs, err := e.env.CommandStore.EnabledCommandIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled command IDs: %w", err)
	}

	idSet := util.IDSet(commandIDs)

	e.RLock()
	defer e.RUnlock()

	for _, app := range e.apps {
		app.RemoveDanglingCommands(idSet)
	}

	return nil
}

func (e *Engine) populateEventListeners(ctx context.Context, lastUpdate time.Time) error {
	queryStart := time.Now()
	listeners, err := e.env.EventListenerStore.EventListenersUpdatedSince(ctx, lastUpdate)
	metrics.ObservePoll("populate_event_listeners", queryStart)
	if err != nil {
		return fmt.Errorf("failed to get event listeners: %w", err)
	}

	var disabled []*model.DeletedEntity
	for _, listener := range listeners {
		if util.CluserForKey(listener.AppID, e.env.Config.ClusterCount) != e.env.Config.ClusterIndex {
			continue
		}

		if !listener.Enabled {
			disabled = append(disabled, &model.DeletedEntity{
				ID:    listener.ID,
				Type:  model.DeletedEntityTypeEventListener,
				AppID: listener.AppID,
			})
			continue
		}

		// Only the first load after startup catches up on missed scheduled
		// runs, later loads are edits or listeners being enabled.
		compiled, err := NewEventListener(listener, e.env, lastUpdate.IsZero())
		if err != nil {
			// NewEventListener already logged the compilation failure.
			continue
		}

		app := e.appForID(listener.AppID)
		app.AddEventListener(listener.ID, compiled)

		if compiled.schedule != nil {
			e.scheduledAppsMu.Lock()
			e.scheduledApps[listener.AppID] = app
			e.scheduledAppsMu.Unlock()
		}
	}

	e.removeEntities(disabled)
	return nil
}

func (e *Engine) removeDanglingEventListeners(ctx context.Context) error {
	listenerIDs, err := e.env.EventListenerStore.EnabledEventListenerIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled event listener IDs: %w", err)
	}

	idSet := util.IDSet(listenerIDs)

	e.RLock()
	defer e.RUnlock()

	for _, app := range e.apps {
		app.RemoveDanglingEventListeners(idSet)
	}

	return nil
}

// removeDeletedEntities drops the entities deleted since the last poll.
// Deleted rows leave nothing for the updated_at polls to find, so the database
// records a tombstone for each.
func (e *Engine) removeDeletedEntities(ctx context.Context, lastUpdate time.Time) error {
	queryStart := time.Now()
	deleted, err := e.env.DeletedEntityStore.DeletedEntitiesSince(ctx, lastUpdate)
	metrics.ObservePoll("populate_deleted_entities", queryStart)
	if err != nil {
		return fmt.Errorf("failed to get deleted entities: %w", err)
	}

	e.removeEntities(deleted)
	return nil
}

// removeEntities drops deleted or disabled entities from their apps. Grouped
// by app, so an app delete or a bulk disable locks each app and rebuilds its
// indexes only once.
func (e *Engine) removeEntities(entities []*model.DeletedEntity) {
	byApp := make(map[string][]*model.DeletedEntity)
	for _, entity := range entities {
		byApp[entity.AppID] = append(byApp[entity.AppID], entity)
	}

	for appID, entities := range byApp {
		if app := e.existingApp(appID); app != nil {
			app.RemoveEntities(entities)
		}
	}
}

func (e *Engine) scheduledEventListeners() []*EventListener {
	e.scheduledAppsMu.Lock()
	defer e.scheduledAppsMu.Unlock()

	var res []*EventListener
	for appID, app := range e.scheduledApps {
		scheduled := app.scheduledEventListeners()
		if len(scheduled) == 0 {
			delete(e.scheduledApps, appID)
			continue
		}
		res = append(res, scheduled...)
	}
	return res
}

// HandleEvent blocks until the event is handled by the corresponding app.
func (e *Engine) HandleEvent(appID string, session *state.State, event gateway.Event) {
	lockStart := time.Now()
	e.RLock()
	app := e.apps[appID]
	e.RUnlock()

	lockDiff := time.Since(lockStart)
	metrics.ObserveLockWait("engine_registry_read", lockDiff)

	if lockDiff > 500*time.Millisecond {
		slog.Warn(
			"Locking engine for handling event took too long",
			slog.String("app_id", appID),
			slog.String("lock_duration", lockDiff.String()),
		)
	}

	if app == nil {
		// Apps are only registered once the engine loads a command, listener or
		// plugin for them, but interactions can target things that live in the
		// database instead: message template buttons and resume points. An app
		// with nothing but message templates would otherwise never see them.
		if _, ok := event.(*gateway.InteractionCreateEvent); ok {
			app = e.appForID(appID)
		} else {
			// The gateway connected before the engine finished loading this
			// app's commands and listeners, or the app has no entities at all.
			// Events dropped here are invisible otherwise, and the window
			// widens with the engine's populate interval.
			metrics.GatewayEventsDropped.Add("unknown_app", 1)
			return
		}
	}

	app.HandleEvent(appID, session, event)
}

type EngineConfig struct {
	MaxStackDepth int
	MaxOperations int
	MaxCredits    int
	ClusterCount  int
	ClusterIndex  int

	PopulateInterval       time.Duration
	RemoveDanglingInterval time.Duration
	PopulateOverlap        time.Duration
}
