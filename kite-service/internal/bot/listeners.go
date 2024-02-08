package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"

	"github.com/merlinfuchs/kite/kite-types/event"
)

func (b *Bot) registerListeners() {
	b.Session.AddHandler(b.handleMessageCreate)
	b.Session.AddHandler(b.handleMessageUpdate)
	b.Session.AddHandler(b.handleGuildCreate)
	b.Session.AddHandler(b.handleGuildUpdate)
}

func (b *Bot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	b.Engine.HandleEvent(context.Background(), &event.Event{
		Type:    event.DiscordMessageCreate,
		GuildID: m.GuildID,
		Data: distype.MessageCreateEvent{
			ID:        distype.Snowflake(m.ID),
			ChannelID: distype.Snowflake(m.ChannelID),
			Content:   m.Content,
		},
	})
}

func (b *Bot) handleMessageUpdate(s *discordgo.Session, m *discordgo.InteractionCreate) {
	var channelID distype.Optional[distype.Snowflake]
	if m.ChannelID != "" {
		v := distype.Snowflake(m.ChannelID)
		channelID = &v
	}

	b.Engine.HandleEvent(context.Background(), &event.Event{
		Type:    event.DiscordInteractionCreate,
		GuildID: m.GuildID,
		Data: distype.InteractionCreateEvent{
			ID:        distype.Snowflake(m.ID),
			Type:      distype.InteractionType(m.Type),
			Token:     m.Token,
			ChannelID: channelID,
			// TODO: Data:      m.Data,
		},
	})
}

func (b *Bot) handleGuildCreate(s *discordgo.Session, m *discordgo.GuildCreate) {
	_, err := b.guildStore.UpsertGuild(context.Background(), model.Guild{
		ID:          m.ID,
		Name:        m.Name,
		Icon:        m.Icon,
		Description: m.Description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to upsert guild from guild create event")
	}
}

func (b *Bot) handleGuildUpdate(s *discordgo.Session, m *discordgo.GuildUpdate) {
	_, err := b.guildStore.UpsertGuild(context.Background(), model.Guild{
		ID:          m.ID,
		Name:        m.Name,
		Icon:        m.Icon,
		Description: m.Description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to upsert guild from guild update event")
	}
}
