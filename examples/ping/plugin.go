package main

import (
	kite "github.com/merlinfuchs/kite/go-sdk"
	"github.com/merlinfuchs/kite/go-sdk/config"
	"github.com/merlinfuchs/kite/go-sdk/discord"
	"github.com/merlinfuchs/kite/go-sdk/log"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/go-types/manifest"
)

func main() {
	config.SetSchema(manifest.ConfigSchema{
		Fields: []manifest.ConfigFieldSchema{
			{
				Name:         "Ping response",
				Description:  "The response to send when a user sends the ping command.",
				Key:          "ping_response",
				Type:         manifest.ConfigFieldTypeString,
				DefaultValue: "Pong!",
			},
		},
	})

	kite.Event(event.DiscordMessageCreate, func(req event.Event) error {
		msg := req.Data.(dismodel.MessageCreateEvent)

		if msg.Content == "!ping" {
			pingResponse := config.String("ping_response")

			_, err := discord.MessageCreate(dismodel.MessageCreateCall{
				ChannelID: msg.ChannelID,
				Content:   pingResponse,
			})
			if err != nil {
				log.Error("Failed to send message: " + err.Error())
				return err
			}
		}

		return nil
	})

	kite.Command("ping", func(i dismodel.Interaction, options []dismodel.ApplicationCommandOptionData) error {
		pingResponse := config.String("ping_response")

		_, err := discord.InteractionResponseCreate(dismodel.InteractionResponseCreateCall{
			ID:    i.ID,
			Token: i.Token,
			Data: dismodel.InteractionResponseData{
				Content: pingResponse,
			},
		})
		if err != nil {
			log.Error("Failed to send message: " + err.Error())
			return err
		}

		return nil
	})
}
