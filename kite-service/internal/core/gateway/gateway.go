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
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

type Gateway struct {
	sync.Mutex

	logStore     store.LogStore
	appStore     store.AppStore
	eventHandler EventHandler

	app     *model.App
	session *state.State
}

func NewGateway(app *model.App, logStore store.LogStore, appStore store.AppStore, eventHandler EventHandler) *Gateway {
	g := &Gateway{
		logStore:     logStore,
		appStore:     appStore,
		eventHandler: eventHandler,
		app:          app,
		session:      createSession(app),
	}

	go g.startGateway()
	return g
}

func (g *Gateway) startGateway() {
	g.Lock()
	defer g.Unlock()

	intents, err := getAppIntents(g.session.Client)
	if err != nil {
		var httpErr *httputil.HTTPError
		if errors.As(err, &httpErr) && httpErr.Status == http.StatusUnauthorized {
			g.createLogEntry(model.LogLevelError, "Discord bot token is invalid, please update it")
			g.disableApp("Discord bot token is invalid, please update it")
			return
		}

		g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to get app intents: %v", err))
		slog.With("error", err).Error("failed to get app intents")
		return
	}

	g.session.AddIntents(intents)

	g.session.AddHandler(func(e gateway.Event) {
		g.eventHandler.HandleEvent(g.app.ID, g.session, e)
	})

	g.session.AddHandler(func(e *gateway.ReadyEvent) {
		g.createLogEntry(model.LogLevelInfo, fmt.Sprintf(
			"Connected to Discord as %s#%s (%s)",
			e.User.Username, e.User.Discriminator, e.User.ID,
		))

		if len(e.Guilds) > 100 {
			g.createLogEntry(model.LogLevelError, "Bots that are in more than 100 servers are currently not supported.")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := g.Close(ctx); err != nil {
				slog.With("error", err).Error("failed to close gateway")
			}

			return
		}
	})

	if err := g.session.Connect(context.TODO()); err != nil {
		// Fatal error, we can't recover
		g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to connect to gateway: %v", err))
		g.disableApp(fmt.Sprintf("Failed to connect to gateway: %v", err))
		return
	}
}

func (g *Gateway) Close(ctx context.Context) error {
	g.Lock()
	defer g.Unlock()

	err := g.session.Close()

	if err != nil && !errors.Is(err, session.ErrClosed) {
		return err
	}

	return nil
}

func (g *Gateway) Update(ctx context.Context, app *model.App) {
	if !app.DiscordStatus.Equals(g.app.DiscordStatus) {
		presence := presenceForApp(app)

		err := g.session.Gateway().Send(ctx, presence)
		if err != nil {
			go g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to update bot status: %v", err))
			slog.With("error", err).Error("failed to send presence update")
		}
	}

	if app.DiscordToken != g.app.DiscordToken {
		g.app = app

		slog.With("app_id", app.ID).Info("Discord token or status changed, closing gateway")
		if err := g.Close(ctx); err != nil {
			slog.With("error", err).Error("failed to close gateway")
		}

		g.session = createSession(app)
		go g.startGateway()
	} else {
		g.app = app
	}
}

func (g *Gateway) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := g.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     g.app.ID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", g.app.ID).Error("Failed to create log entry from gateway")
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
		slog.With("error", err).With("app_id", g.app.ID).Error("Failed to disable app from gateway")
	}
}
