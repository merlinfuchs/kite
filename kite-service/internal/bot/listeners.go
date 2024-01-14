package bot

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/event"
)

func (b *Bot) registerListeners() {
	b.Session.AddHandler(b.handleMessageCreate)
	b.Session.AddHandler(b.handleMessageUpdate)
}

func (b *Bot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	err := b.Engine.HandleEvent(context.Background(), &event.Event{
		Type:    event.DiscordMessageCreate,
		GuildID: m.GuildID,
		Data: dismodel.MessageCreateEvent{
			ID:        m.ID,
			ChannelID: m.ChannelID,
			Content:   m.Content,
		},
	})
	if err != nil {
		fmt.Printf("failed to handle event: %v\n", err)
	}
}

func (b *Bot) handleMessageUpdate(s *discordgo.Session, m *discordgo.InteractionCreate) {
	err := b.Engine.HandleEvent(context.Background(), &event.Event{
		Type:    event.DiscordInteractionCreate,
		GuildID: m.GuildID,
		Data: dismodel.InteractionCreateEvent{
			ID:        m.ID,
			Type:      dismodel.InteractionType(m.Type),
			Token:     m.Token,
			ChannelID: m.ChannelID,
			Data:      m.Data,
		},
	})
	if err != nil {
		fmt.Printf("failed to handle event: %v\n", err)
	}
}
