package main

import (
	kite "github.com/merlinfuchs/kite/go-sdk"
	"github.com/merlinfuchs/kite/go-sdk/discord"
	"github.com/merlinfuchs/kite/go-sdk/log"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/event"
)

func init() {
	log.Debug("Plugin loaded")

	kite.Event(event.DiscordMessageCreate, func(req event.Event) error {
		msg := req.Data.(dismodel.MessageCreateEvent)

		channel, err := discord.ChannelGet(msg.ChannelID)
		if err != nil {
			log.Error("Failed to get channel: " + err.Error())
			return err
		}

		if msg.Content == "!ping" {
			_, err := discord.MessageCreate(dismodel.MessageCreateCall{
				ChannelID: msg.ChannelID,
				Content:   "Pong! " + channel.Name,
			})
			if err != nil {
				log.Error("Failed to send message: " + err.Error())
				return err
			}
		}

		return nil
	})

	kite.Command("ping", func(i dismodel.Interaction, options []dismodel.ApplicationCommandOptionData) error {
		_, err := discord.InteractionResponseCreate(dismodel.InteractionResponseCreateCall{
			ID:    i.ID,
			Token: i.Token,
			Data: dismodel.InteractionResponseData{
				Content: "Pong!",
			},
		})
		if err != nil {
			log.Error("Failed to send message: " + err.Error())
			return err
		}

		return nil
	})
}

func main() {}
