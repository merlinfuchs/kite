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

// statusRotationInterval is how often a gateway with status rotation enabled
// cycles to the next configured Discord status.
const statusRotationInterval = time.Minute

type Gateway struct {
	logStore       store.LogStore
	appStore       store.AppStore
	planManager    *plan.PlanManager
	eventHandler   EventHandler
	tokenCrypt     *util.SymmetricCrypt
	pluginRegistry *plugin.Registry

	app     *model.App
	session *state.State

	// statusRotateMu guards statusRotateCancel, which stops the currently
	// running status-rotation goroutine (if any) started by
	// applyStatusRotation.
	statusRotateMu     sync.Mutex
	statusRotateCancel context.CancelFunc

	// intents is what this connection identified with, or zero before it has
	// been computed. A computed set always includes IntentGuilds, so zero is
	// unambiguous. Compared against a freshly computed set on refresh; a
	// change requires a reconnect, since intents are fixed at IDENTIFY.
	intents gateway.Intents

	ctx    context.Context
	cancel context.CancelFunc
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
		app:            app,
		session:        session,
	}

	g.ctx, g.cancel = context.WithCancel(context.Background())

	go g.startGateway()
	g.applyStatusRotation(app)
	return g, nil
}

func (g *Gateway) startGateway() {
	intents, err := g.computeIntents(g.ctx)
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
			slog.String("app_id", g.app.ID),
			slog.String("error", err.Error()),
		)
		return
	}

	g.intents = intents
	g.session.AddIntents(intents)

	slog.Debug(
		"Computed gateway intents",
		slog.String("app_id", g.app.ID),
		slog.Uint64("intents", uint64(intents)),
	)

	g.session.AddHandler(func(e gateway.Event) {
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
		g.eventHandler.HandleEvent(g.app.ID, g.session, e)
	})

	g.session.AddHandler(func(e *gateway.ReadyEvent) {
		slog.Info(
			"Received ready event",
			slog.String("app_id", g.app.ID),
			slog.String("user_id", e.User.ID.String()),
			slog.String("username", e.User.Username),
			slog.Int("guilds", len(e.Guilds)),
		)
		g.createLogEntry(model.LogLevelInfo, fmt.Sprintf(
			"Connected to Discord as %s#%s (%s)",
			e.User.Username, e.User.Discriminator, e.User.ID,
		))

		features := g.planManager.AppFeatures(g.ctx, g.app.ID)
		if len(e.Guilds) > features.MaxGuilds {
			g.createLogEntry(model.LogLevelError, "Bots that are in more than 100 servers are currently not supported.")
			g.disableApp("Bots that are in more than 100 servers are currently not supported.")
			return
		}
	})

	if err := g.session.Connect(g.ctx); err != nil {
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
func (g *Gateway) computeIntents(ctx context.Context) (gateway.Intents, error) {
	app, err := g.session.Client.CurrentApplication()
	if err != nil {
		return 0, fmt.Errorf("failed to get current application: %w", err)
	}

	reqs, err := g.appRequirements(ctx)
	if err != nil {
		slog.Error(
			"Failed to load gateway requirements, falling back to all permitted intents",
			slog.String("app_id", g.app.ID),
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

	row, err := g.appStore.AppGatewayRequirements(ctx, g.app.ID)
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
	g.cancel()
	err := g.session.Close()

	if err != nil && !errors.Is(err, session.ErrClosed) {
		return fmt.Errorf("failed to close gateway: %w", err)
	}

	return nil
}

func (g *Gateway) Update(ctx context.Context, app *model.App) {
	if !app.DiscordStatus.Equals(g.app.DiscordStatus) {
		presence := presenceForApp(app)

		err := g.session.Gateway().Send(ctx, presence)
		if err != nil {
			go g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to update bot status: %v", err))
			slog.Error(
				"Failed to send presence update",
				slog.String("app_id", app.ID),
				slog.String("error", err.Error()),
			)
		}

		g.applyStatusRotation(app)
	}

	if app.DiscordToken != g.app.DiscordToken {
		g.app = app

		slog.Info(
			"Discord token changed, reconnecting gateway",
			slog.String("app_id", app.ID),
		)
		g.restart()
		return
	}

	g.app = app
}

// applyStatusRotation (re)starts the background goroutine that cycles the
// app's presence through every configured status once a minute, replacing
// whatever rotation was previously running. If rotation is disabled, or
// there are fewer than two statuses to rotate between, any existing rotation
// is stopped and nothing new is started -- presenceForApp already shows the
// single active status in that case.
func (g *Gateway) applyStatusRotation(app *model.App) {
	g.statusRotateMu.Lock()
	defer g.statusRotateMu.Unlock()

	if g.statusRotateCancel != nil {
		g.statusRotateCancel()
		g.statusRotateCancel = nil
	}

	if app.DiscordStatus == nil || !app.DiscordStatus.RotateEnabled || len(app.DiscordStatus.Statuses) < 2 {
		return
	}

	// Copy the slice so later config changes (which call applyStatusRotation
	// again with a fresh app) can't race with this goroutine reading it.
	statuses := make([]model.AppDiscordStatusEntry, len(app.DiscordStatus.Statuses))
	copy(statuses, app.DiscordStatus.Statuses)

	ctx, cancel := context.WithCancel(g.ctx)
	g.statusRotateCancel = cancel

	go func() {
		ticker := time.NewTicker(statusRotationInterval)
		defer ticker.Stop()

		index := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				index = (index + 1) % len(statuses)
				entry := statuses[index]

				if err := g.session.Gateway().Send(ctx, presenceForStatusEntry(&entry)); err != nil {
					slog.Error(
						"Failed to send rotating presence update",
						slog.String("app_id", g.app.ID),
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}()
}

// RefreshIntents recomputes the app's required intents and reconnects if they
// changed. Intents are fixed at IDENTIFY, so a reconnect is the only way to
// apply a change.
//
// Called when an app's event listeners or plugin instances change. Any error
// leaves the connection alone: the current intent set was correct as of the
// last computation, so keeping it beats a reconnect loop.
func (g *Gateway) RefreshIntents(ctx context.Context) {
	if g.intents == 0 {
		// Still starting up; startGateway will compute the current set.
		return
	}

	intents, err := g.computeIntents(ctx)
	if err != nil {
		slog.Error(
			"Failed to compute intents while refreshing",
			slog.String("app_id", g.app.ID),
			slog.String("error", err.Error()),
		)
		return
	}

	if intents == g.intents {
		return
	}

	slog.Info(
		"Gateway intents changed, reconnecting",
		slog.String("app_id", g.app.ID),
		slog.Uint64("old_intents", uint64(g.intents)),
		slog.Uint64("new_intents", uint64(intents)),
	)
	metrics.GatewayIntentReconnects.Add(1)

	g.restart()
}

// restart tears the connection down and brings it back up with freshly
// computed intents.
func (g *Gateway) restart() {
	if err := g.Close(); err != nil {
		slog.Error(
			"Failed to close gateway",
			slog.String("error", err.Error()),
			slog.String("app_id", g.app.ID),
		)
	}

	session, err := createSession(g.tokenCrypt, g.app)
	if err != nil {
		g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to create session: %v", err))
		return
	}

	// Close cancelled the context. Without a fresh one, Connect returns
	// immediately with a context error and startGateway treats that as fatal
	// and disables the app.
	g.ctx, g.cancel = context.WithCancel(context.Background())
	g.session = session
	g.intents = 0
	go g.startGateway()
	g.applyStatusRotation(g.app)
}

func (g *Gateway) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := g.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     g.app.ID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to create log entry from gateway",
			slog.String("error", err.Error()),
			slog.String("app_id", g.app.ID),
		)
	}
}

func (g *Gateway) disableApp(reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := g.appStore.DisableApp(ctx, store.AppDisableOpts{
		ID:             g.app.ID,
		DisabledReason: null.StringFrom(reason),
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to disable app from gateway",
			slog.String("error", err.Error()),
			slog.String("app_id", g.app.ID),
		)
	}
}
