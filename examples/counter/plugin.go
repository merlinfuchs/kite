package main

import (
	"strconv"

	kite "github.com/merlinfuchs/kite/go-sdk"
	"github.com/merlinfuchs/kite/go-sdk/discord"
	"github.com/merlinfuchs/kite/go-sdk/kv"
	"github.com/merlinfuchs/kite/go-sdk/log"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/event"
)

func main() {
	log.Debug("Plugin loaded")

	kite.Event(event.DiscordMessageCreate, func(req event.Event) error {
		msg := req.Data.(dismodel.MessageCreateEvent)
		store := kv.New()

		count, err := strconv.Atoi(msg.Content)

		if err == nil {
			counter, err := store.Increase(msg.ChannelID, 1)
			if err != nil {
				log.Error("Failed to increase counter: " + err.Error())
				return err
			}

			if count != counter.Int() {
				_, err = store.Delete(msg.ChannelID)
				if err != nil {
					log.Error("Failed to delete counter: " + err.Error())
					return err
				}

				_, err := discord.MessageCreate(dismodel.MessageCreateCall{
					ChannelID: msg.ChannelID,
					Content:   "Wrong counter value! The counter has been reset.",
				})
				if err != nil {
					log.Error("Failed to send message: " + err.Error())
					return err
				}
			}
		}

		return nil
	})
}
