package main

import (
	kite "github.com/merlinfuchs/kite/kite-sdk-go"
	"github.com/merlinfuchs/kite/kite-sdk-go/command"
	"github.com/merlinfuchs/kite/kite-sdk-go/discord"
	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
	"github.com/merlinfuchs/kite/kite-sdk-go/log"

	"strings"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/event"
)

func main() {
	kite.Event(event.DiscordMessageCreate, func(req event.Event) error {
		msg := req.Data.(distype.MessageCreateEvent)
		store := kite.KV()

		if strings.HasPrefix(msg.Content, "!echo") {
			echoResponse := strings.TrimPrefix(msg.Content, "!echo")

			newMsg, err := discord.MessageCreate(distype.MessageCreateRequest{
				ChannelID: msg.ChannelID,
				MessageCreateParams: distype.MessageCreateParams{
					Content: &echoResponse,
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
		store := kite.KV()

		if strings.HasPrefix(msg.Content, "!echo") {
			echoResponse := strings.TrimPrefix(msg.Content, "!echo")

			responseMessageID, err := store.Get(string(msg.ID))
			if err != nil {
				log.Error("Failed to get key: " + err.Error())
				return err
			}

			_, err = discord.MessageUpdate(distype.MessageEditRequest{
				ChannelID: msg.ChannelID,
				MessageID: distype.Snowflake(responseMessageID.String()),
				MessageEditParams: distype.MessageEditParams{
					Content: &echoResponse,
				},
			})
			if err != nil {
				log.Error("Failed to edit message: " + err.Error())
				return err
			}
		}

		return nil
	})

	kite.Command("echo").
		WithDescription("Echo the text you provided!").
		WithOption(command.CommandOption{
			Type:        distype.ApplicationCommandOptionTypeString,
			Name:        "text",
			Description: "The text to echo.",
		}).
		WithHandler(func(i distype.Interaction, options []distype.ApplicationCommandDataOption) error {
			text := options[0].Value.(string)

			_, err := discord.InteractionResponseCreate(distype.InteractionResponseCreateRequest{
				InteractionID:    i.ID,
				InteractionToken: i.Token,
				InteractionResponse: distype.InteractionResponse{
					Type: distype.InteractionResponseTypeChannelMessageWithSource,
					Data: &distype.InteractionMessageResponse{
						Content: &text,
					},
				},
			})
			if err != nil {
				log.Error("Failed to send message: " + err.Error())
				return err
			}

			return nil
		})
}
