package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
)

type Bot struct {
	Session *discordgo.Session
	Engine  *engine.PluginEngine
}

func New(token string) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	session.Identify.Presence = discordgo.GatewayStatusUpdate{
		Game: discordgo.Activity{
			Name:  "kite.bot",
			State: "kite.bot",
			Type:  discordgo.ActivityTypeCustom,
		},
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready!")
	})

	b := &Bot{
		Session: session,
	}

	b.registerListeners()

	return b, nil
}

func (b *Bot) Start() error {
	return b.Session.Open()
}
