package main

import (
	kite "github.com/merlinfuchs/kite/kite-sdk-go"
	"github.com/merlinfuchs/kite/kite-sdk-go/config"
	"github.com/merlinfuchs/kite/kite-sdk-go/discord"
	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
	"github.com/merlinfuchs/kite/kite-sdk-go/log"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/event"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
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
		msg := req.Data.(distype.MessageCreateEvent)
		store := kite.KV()

		if msg.Content == "!ping" {
			pingResponse := config.String("ping_response")

			newMsg, err := discord.MessageCreate(distype.MessageCreateRequest{
				ChannelID: msg.ChannelID,
				MessageCreateParams: distype.MessageCreateParams{
					Content: &pingResponse,
				},
			})
			if err != nil {
				log.Error("Failed to send message: " + err.Error())
				return err
			}

			err = store.Set(string(msg.ID), kv.KVString(newMsg.ID))
			if err != nil {
				log.Error("Failed to set key: " + err.Error())
				return err
			}
		}

		return nil
	})

	kite.Event(event.DiscordMessageUpdate, func(req event.Event) error {
		msg := req.Data.(distype.MessageUpdateEvent)

		if msg.Content == "!ping" {
			pingResponse := config.String("ping_response")

			_, err := discord.MessageCreate(distype.MessageCreateRequest{
				ChannelID: msg.ChannelID,
				MessageCreateParams: distype.MessageCreateParams{
					Content: &pingResponse,
				},
			})
			if err != nil {
				log.Error("Failed to send message: " + err.Error())
				return err
			}
		}

		return nil
	})

	kite.Command("ping").
		WithDescription("Replies with pong!").
		WithHandler(func(i distype.Interaction, options []distype.ApplicationCommandDataOption) error {
			pingResponse := config.String("ping_response")

			_, err := discord.InteractionResponseCreate(distype.InteractionResponseCreateRequest{
				InteractionID:    i.ID,
				InteractionToken: i.Token,
				InteractionResponse: distype.InteractionResponse{
					Type: distype.InteractionResponseTypeChannelMessageWithSource,
					Data: &distype.InteractionMessageCreateResponse{
						Content: &pingResponse,
					},
				},
			})
			if err != nil {
				log.Error("Failed to respond to interaction: " + err.Error())
				return err
			}

			return nil
		})
}
