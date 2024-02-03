package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/kite-service/internal/bot/state"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type Bot struct {
	Session    *discordgo.Session
	Engine     *engine.Engine
	State      *state.BotState
	guildStore store.GuildStore
}

func New(token string, pg *postgres.Client) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildMessages) // | discordgo.IntentsGuildMembers)
	session.Identify.Presence = discordgo.GatewayStatusUpdate{
		Game: discordgo.Activity{
			Name:  "kite.onl",
			State: "kite.onl",
			Type:  discordgo.ActivityTypeCustom,
		},
	}

	session.StateEnabled = true

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready!")
	})

	b := &Bot{
		Session:    session,
		State:      state.New(session.State, session),
		guildStore: pg,
	}

	b.registerListeners()

	return b, nil
}

func (b *Bot) Start() error {
	return b.Session.Open()
}
