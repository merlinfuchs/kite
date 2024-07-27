package gateway

import (
	"context"
	"log/slog"
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-common/model"
)

type Gateway struct {
	sync.Mutex

	eventHandler EventHandler
	app          *model.App
	shards       *shard.Manager
}

func NewGateway(app *model.App, eventHandler EventHandler) *Gateway {
	g := &Gateway{
		eventHandler: eventHandler,
		app:          app,
	}

	go g.startGateway()
	return g
}

func (g *Gateway) startGateway() {
	g.Lock()
	defer g.Unlock()

	newShard := state.NewShardFunc(func(m *shard.Manager, s *state.State) {
		s.AddIntents(gateway.IntentGuilds | gateway.IntentGuildMessages | gateway.IntentMessageContent)

		s.AddHandler(func(e gateway.Event) {
			g.eventHandler.HandleEvent(g.app.ID, e)
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
					State: "🪁 kite.onl",
				},
			},
		},
	}, newShard)
	if err != nil {
		// TODO: create log entry
		slog.With("error", err).Error("failed to create shard manager")
		return
	}

	if err := shards.Open(context.TODO()); err != nil {
		// TODO: create log entry
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
		if err := g.Close(ctx); err != nil {
			slog.With("error", err).Error("failed to close gateway")
		}

		go g.startGateway()
	}

	g.app = app
}
