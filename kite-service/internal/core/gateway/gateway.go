package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type Gateway struct {
	sync.Mutex

	logStore     store.LogStore
	eventHandler EventHandler

	app    *model.App
	shards *shard.Manager
}

func NewGateway(app *model.App, logStore store.LogStore, eventHandler EventHandler) *Gateway {
	g := &Gateway{
		logStore:     logStore,
		eventHandler: eventHandler,
		app:          app,
	}

	go g.startGateway()
	return g
}

func (g *Gateway) startGateway() {
	g.Lock()
	defer g.Unlock()

	client := api.NewClient("Bot " + g.app.DiscordToken)
	intents, err := getAppIntents(client)
	if err != nil {
		go g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to get app intents: %v", err))
		slog.With("error", err).Error("failed to get app intents")
		return
	}

	newShard := state.NewShardFunc(func(m *shard.Manager, s *state.State) {
		s.AddIntents(intents)

		s.AddHandler(func(e gateway.Event) {
			g.eventHandler.HandleEvent(g.app.ID, e)
		})

		s.AddHandler(func(e *gateway.ReadyEvent) {
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
	})

	shards, err := shard.NewIdentifiedManager(gateway.IdentifyCommand{
		Token: "Bot " + g.app.DiscordToken,
		Presence: &gateway.UpdatePresenceCommand{
			Status: discord.OnlineStatus,
			Activities: []discord.Activity{
				{
					Type:  discord.CustomActivity,
					Name:  "kite.onl",
					State: "🪁 Powered by Kite.onl",
				},
			},
		},
	}, newShard)
	if err != nil {
		go g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to create shard manager, make sure the token is correct: %v", err))
		slog.With("error", err).Error("failed to create shard manager")
		return
	}

	if shards.NumShards() > 1 {
		go g.createLogEntry(model.LogLevelError, "Sharding is not supported, your bot is in more than 1000 servers.")
		slog.With("error", err).Error("sharding is not supported")
		return
	}

	if err := shards.Open(context.TODO()); err != nil {
		go g.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to open shard manager: %v", err))
		slog.With("error", err).Error("failed to open shard manager")
		return
	}

	g.shards = shards
}

func (g *Gateway) Close(ctx context.Context) error {
	g.Lock()
	defer g.Unlock()

	if g.shards != nil {
		if err := g.shards.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Gateway) Update(ctx context.Context, app *model.App) {
	if app.DiscordToken != g.app.DiscordToken {
		slog.With("app_id", app.ID).Info("Discord token changed, closing gateway")
		if err := g.Close(ctx); err != nil {
			slog.With("error", err).Error("failed to close gateway")
		}

		go g.startGateway()
	}

	g.app = app
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
