package main

import (
	"strconv"

	kite "github.com/merlinfuchs/kite/go-sdk"
	"github.com/merlinfuchs/kite/go-sdk/discord"
	"github.com/merlinfuchs/kite/go-sdk/kv"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/event"
)

const resetMessage = "Wrong counter value! The counter has been reset."

func main() {
	kite.Event(event.DiscordMessageCreate, handleMessageCreateEvent)
}

func handleMessageCreateEvent(req event.Event) error {
	msg := req.Data.(dismodel.MessageCreateEvent)

	if count, err := strconv.Atoi(msg.Content); err == nil {
		if err := updateCounter(msg.ChannelID, count); err != nil {
			return err
		}
	}

	return nil
}

func updateCounter(channelID string, count int) error {
	store := kv.New()

	counter, err := store.Increase(channelID, 1)
	if err != nil {
		return err
	}

	if count != counter.Int() {
		if _, err := store.Delete(channelID); err != nil {
			return err
		}

		if _, err := discord.MessageCreate(dismodel.MessageCreateCall{
			ChannelID: channelID,
			Content:   resetMessage,
		}); err != nil {
			return err
		}
	}

	return nil
}
