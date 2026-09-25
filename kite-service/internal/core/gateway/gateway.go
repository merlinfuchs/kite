package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/metrics"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"gopkg.in/guregu/null.v4"
)

type Gateway struct {
	logStore       store.LogStore
	appStore       store.AppStore
	planManager    *plan.PlanManager
	eventHandler   EventHandler
	tokenCrypt     *util.SymmetricCrypt
	pluginRegistry *plugin.Registry

	// appID never changes, unlike the rest of app.
	appID string

	// mu guards the app and the connection, which Update and restart replace
	// while the manager, the API and the engine's scheduler read them.
	mu      sync.RWMutex
	app     *model.App
	session *state.State
	// intents is what this connection identified with, or zero before it has
	// been computed. A computed set always includes IntentGuilds, so zero is
	// unambiguous. Compared against a freshly computed set on refresh; a
	// change requires a reconnect, since intents are fixed at IDENTIFY.
	intents gateway.Intents
	// ctx is cancelled when the connection is replaced or closed, so the
	// goroutine that started it knows it's no longer wanted.
	ctx    context.Context
	cancel context.CancelFunc
	// closed is set once the manager closed the gateway, after which it must
	// not reconnect.
	closed bool

	// rotationEntryID is the status entry last shown by rotatePresence, or
	// empty if the app isn't rotating. Only accessed by the manager's loop.
	rotationEntryID string
}

func NewGateway(
	app *model.App,
	logStore store.LogStore,
	appStore store.AppStore,
	planManager *plan.PlanManager,
	eventHandler EventHandler,
	tokenCrypt *util.SymmetricCrypt,
	pluginRegistry *plugin.Registry,
) (*Gateway, error) {
	session, err := createSession(tokenCrypt, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	g := &Gateway{
		logStore:       logStore,
		appStore:       appStore,
		planManager:    planManager,
		eventHandler:   eventHandler,
		tokenCrypt:     tokenCrypt,
		pluginRegistry: pluginRegistry,
		appID:          app.ID,
		app:            app,
		session:        session,
	}

	g.ctx, g.cancel = context.WithCancel(context.Background())

	go g.startGateway(session, g.ctx)
	return g, nil
}

// Session returns the current connection's session.
func (g *Gateway) Session() *state.State {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.session
}

func (g *Gateway) currentApp() *model.App {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.app
}

// startGateway connects session. It's passed in rather than read from g, so a
// connection that restart already replaced never touches the new one. Once
// ctx is cancelled the connection was replaced or closed on purpose, so its
// errors don't disable the app.
func (g *Gateway) startGateway(session *state.State, ctx context.Context) {
	intents, err := g.computeIntents(ctx, session)
	if ctx.Err() != nil {
		return
	}
	if err != nil {
		var httpErr *httputil.HTTPError
		if errors.As(err, &httpErr) && httpErr.Status == http.StatusUnauthorized {
			g.createLogEntry(model.LogLevelError, "Discord bot token is invalid, please update it")
			g.disableApp("Discord bot token is invalid, please update it")
			return
		}

		g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to get app intents: %v", err))
		slog.Error(
			"Failed to get app intents",
			slog.String("app_id", g.appID),
			slog.String("error", err.Error()),
		)
		return
	}

	g.mu.Lock()
	if g.session == session {
		g.intents = intents
	}
	g.mu.Unlock()
	session.AddIntents(intents)

	slog.Debug(
		"Computed gateway intents",
		slog.String("app_id", g.appID),
		slog.Uint64("intents", uint64(intents)),
	)

	session.AddHandler(func(e gateway.Event) {
		// Protocol frames -- heartbeat acks, hello, reconnect, invalid session
		// -- report an empty event type. Nothing downstream can ever match
		// them: no event listener type and no plugin event type is empty. At
		// steady state they are the single largest source of events, roughly
		// one heartbeat per connection every 41s, so dropping them here keeps
		// them off the dispatch path entirely and stops them dominating both
		// the event counter and the dropped-event counter.
		eventType := e.EventType()
		if eventType == "" {
			return
		}

		metrics.GatewayEvents.Add(string(eventType), 1)
		g.eventHandler.HandleEvent(g.appID, session, e)
	})

	session.AddHandler(func(e *gateway.ReadyEvent) {
		slog.Info(
			"Received ready event",
			slog.String("app_id", g.appID),
			slog.String("user_id", e.User.ID.String()),
			slog.String("username", e.User.Username),
			slog.Int("guilds", len(e.Guilds)),
		)
		g.createLogEntry(model.LogLevelInfo, fmt.Sprintf(
			"Connected to Discord as %s#%s (%s)",
			e.User.Username, e.User.Discriminator, e.User.ID,
		))

		features := g.planManager.AppFeatures(ctx, g.appID)
		if len(e.Guilds) > features.MaxGuilds && ctx.Err() == nil {
			g.createLogEntry(model.LogLevelError, "Bots that are in more than 100 servers are currently not supported.")
			g.disableApp("Bots that are in more than 100 servers are currently not supported.")
			return
		}
	})

	if err := session.Connect(ctx); err != nil && ctx.Err() == nil {
		// Fatal error, we can't recover
		g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to connect to gateway: %v", err))
		g.disableApp(fmt.Sprintf("Failed to connect to gateway: %v", err))
		return
	}
}

// computeIntents derives the intent set this app should identify with.
//
// The returned error is always from fetching the application, so callers can
// still inspect it for a 401. A failure to load requirements is not fatal: it
// falls back to every intent the app is permitted, because failing closed
// would silently stop delivering events.
func (g *Gateway) computeIntents(ctx context.Context, session *state.State) (gateway.Intents, error) {
	app, err := session.Client.CurrentApplication()
	if err != nil {
		return 0, fmt.Errorf("failed to get current application: %w", err)
	}

	reqs, err := g.appRequirements(ctx)
	if err != nil {
		slog.Error(
			"Failed to load gateway requirements, falling back to all permitted intents",
			slog.String("app_id", g.appID),
			slog.String("error", err.Error()),
		)
		return allPermittedIntents(app.Flags), nil
	}

	return intentsForRequirements(reqs, app.Flags), nil
}

// appRequirements loads what this app consumes from the gateway and resolves
// its plugin resources to concrete event types.
func (g *Gateway) appRequirements(ctx context.Context) (model.AppGatewayRequirements, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row, err := g.appStore.AppGatewayRequirements(ctx, g.appID)
	if err != nil {
		return model.AppGatewayRequirements{}, fmt.Errorf("failed to get gateway requirements: %w", err)
	}

	return model.AppGatewayRequirements{
		EventListenerTypes:  row.EventListenerTypes,
		PluginEventTypes:    g.pluginRegistry.EventTypesForResources(row.PluginResources),
		HasMessageInstances: row.HasMessageInstances,
	}, nil
}

func (g *Gateway) Close() error {
	g.mu.Lock()
	g.closed = true
	old, cancel := g.session, g.cancel
	g.mu.Unlock()

	cancel()
	return closeSession(old)
}

// closeSession can block while the websocket closes, so it's never called
// with g.mu held.
func closeSession(s *state.State) error {
	err := s.Close()
	if err != nil && !errors.Is(err, session.ErrClosed) {
		return fmt.Errorf("failed to close gateway: %w", err)
	}
	return nil
}

// errNotConnected is returned for commands sent before the session opened its
// gateway connection, e.g. right after a restart.
var errNotConnected = errors.New("gateway isn't connected yet")

func (g *Gateway) sendPresence(ctx context.Context, presence *gateway.UpdatePresenceCommand) error {
	gw := g.Session().Gateway()
	if gw == nil {
		return errNotConnected
	}
	return gw.Send(ctx, presence)
}

func (g *Gateway) Update(ctx context.Context, app *model.App) {
	g.mu.Lock()
	old := g.app
	g.app = app
	g.mu.Unlock()

	if !app.DiscordStatus.Equals(old.DiscordStatus) {
		// A session that isn't connected yet identifies with the new status.
		err := g.sendPresence(ctx, presenceForApp(app))
		if err != nil && !errors.Is(err, errNotConnected) {
			go g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to update bot status: %v", err))
			slog.Error(
				"Failed to send presence update",
				slog.String("app_id", app.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	if app.DiscordToken != old.DiscordToken {
		slog.Info(
			"Discord token changed, reconnecting gateway",
			slog.String("app_id", app.ID),
		)
		g.restart(nil)
	}
}

// rotatePresence shows the rotation entry for the given time. When rotation is
// off or not allowed, it goes back to the active status if it was rotating.
func (g *Gateway) rotatePresence(ctx context.Context, now time.Time, allowed bool) {
	app := g.currentApp()
	status := app.DiscordStatus

	var presence *gateway.UpdatePresenceCommand
	if allowed && status.Rotates() {
		entry := status.RotationEntry(now)
		if entry.ID == g.rotationEntryID {
			return
		}
		g.rotationEntryID = entry.ID
		presence = presenceForStatusEntry(entry)
	} else if g.rotationEntryID != "" {
		g.rotationEntryID = ""
		presence = presenceForApp(app)
	} else {
		return
	}

	if err := g.sendPresence(ctx, presence); err != nil {
		// Sent again on the next tick.
		g.rotationEntryID = ""
		if errors.Is(err, errNotConnected) {
			return
		}
		slog.Error(
			"Failed to send rotating presence update",
			slog.String("app_id", g.appID),
			slog.String("error", err.Error()),
		)
	}
}

// RefreshIntents recomputes the app's required intents and reconnects if they
// changed. Intents are fixed at IDENTIFY, so a reconnect is the only way to
// apply a change.
//
// Called when an app's event listeners or plugin instances change. Any error
// leaves the connection alone: the current intent set was correct as of the
// last computation, so keeping it beats a reconnect loop.
func (g *Gateway) RefreshIntents(ctx context.Context) {
	g.mu.RLock()
	session, current := g.session, g.intents
	g.mu.RUnlock()

	if current == 0 {
		// Still starting up; startGateway will compute the current set.
		return
	}

	intents, err := g.computeIntents(ctx, session)
	if err != nil {
		slog.Error(
			"Failed to compute intents while refreshing",
			slog.String("app_id", g.appID),
			slog.String("error", err.Error()),
		)
		return
	}

	if intents == current {
		return
	}

	slog.Info(
		"Gateway intents changed, reconnecting",
		slog.String("app_id", g.appID),
		slog.Uint64("old_intents", uint64(current)),
		slog.Uint64("new_intents", uint64(intents)),
	)
	metrics.GatewayIntentReconnects.Add(1)

	// Skipped if something else restarted the gateway in the meantime, its
	// connection already computes the current intents.
	g.restart(session)
}

// restart tears the connection down and brings it back up with freshly
// computed intents. If expected is set, it only restarts while that session is
// still the current one.
func (g *Gateway) restart(expected *state.State) {
	session, ctx, err := g.replaceSession(expected)
	if err != nil {
		g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to create session: %v", err))
		return
	}
	if session == nil {
		return
	}

	go g.startGateway(session, ctx)
}

// replaceSession swaps in a new, not yet connected session and closes the old
// one. It returns a nil session if the gateway was closed, or expected is set
// and no longer the current session.
func (g *Gateway) replaceSession(expected *state.State) (*state.State, context.Context, error) {
	g.mu.Lock()
	if g.closed || (expected != nil && g.session != expected) {
		g.mu.Unlock()
		return nil, nil, nil
	}

	session, err := createSession(g.tokenCrypt, g.app)
	if err != nil {
		g.mu.Unlock()
		return nil, nil, err
	}

	old, cancel := g.session, g.cancel
	g.ctx, g.cancel = context.WithCancel(context.Background())
	g.session = session
	g.intents = 0
	ctx := g.ctx
	g.mu.Unlock()

	cancel()
	if err := closeSession(old); err != nil {
		slog.Error(
			"Failed to close gateway",
			slog.String("error", err.Error()),
			slog.String("app_id", g.appID),
		)
	}

	return session, ctx, nil
}

func (g *Gateway) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := g.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     g.appID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to create log entry from gateway",
			slog.String("error", err.Error()),
			slog.String("app_id", g.appID),
		)
	}
}

func (g *Gateway) disableApp(reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := g.appStore.DisableApp(ctx, store.AppDisableOpts{
		ID:             g.appID,
		DisabledReason: null.StringFrom(reason),
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to disable app from gateway",
			slog.String("error", err.Error()),
			slog.String("app_id", g.appID),
		)
	}
}
