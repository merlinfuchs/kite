package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/merlinfuchs/dismod/distype"

	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-types/event"
)

func (b *Bot) registerListeners() {
	b.Cluster.AddEventListener(distype.EventTypeReady, b.handleReady)
	b.Cluster.AddEventListener(distype.EventTypeGuildCreate, b.handleGuildCreate)
	b.Cluster.AddEventListener(distype.EventTypeGuildUpdate, b.handleGuildUpdate)
	b.Cluster.AddEventListener(distype.EventTypeMessageCreate, b.handleMessageCreate)
	b.Cluster.AddEventListener(distype.EventTypeInteractionCreate, b.handleInteractionCreate)
}

func (b *Bot) handleReady(s int, e interface{}) {
	slog.Info("Shard is ready", "shard_id", s)
}

func (b *Bot) handleGuildCreate(s int, e interface{}) {
	g := e.(*distype.GuildCreateEvent)

	_, err := b.guildStore.UpsertGuild(context.Background(), model.Guild{
		ID:          g.ID,
		Name:        g.Name,
		Icon:        g.Icon.Value,
		Description: g.Description.Value,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to upsert guild from guild create event")
	}
}

func (b *Bot) handleGuildUpdate(s int, e interface{}) {
	g := e.(*distype.GuildUpdateEvent)

	_, err := b.guildStore.UpsertGuild(context.Background(), model.Guild{
		ID:          g.ID,
		Name:        g.Name,
		Icon:        g.Icon.Value,
		Description: g.Description.Value,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to upsert guild from guild update event")
	}
}

func (b *Bot) handleMessageCreate(s int, e interface{}) {
	m := e.(*distype.MessageCreateEvent)

	if m.GuildID == nil {
		return
	}

	b.Engine.HandleEvent(context.Background(), &event.Event{
		Type:    event.DiscordMessageCreate,
		GuildID: *m.GuildID,
		Data:    m,
	})
}

func (b *Bot) handleInteractionCreate(s int, e interface{}) {
	i := e.(*distype.InteractionCreateEvent)

	if i.GuildID == nil {
		return
	}

	b.Engine.HandleEvent(context.Background(), &event.Event{
		Type:    event.DiscordInteractionCreate,
		GuildID: *i.GuildID,
		Data:    i,
	})
}
