package main

import (
	"strconv"

	"github.com/merlinfuchs/dismod/distype"
	kite "github.com/merlinfuchs/kite/kite-sdk-go"
	"github.com/merlinfuchs/kite/kite-sdk-go/discord"
	"github.com/merlinfuchs/kite/kite-sdk-go/kv"

	"github.com/merlinfuchs/kite/kite-types/event"
)

var resetMessage = "Wrong counter value! The counter has been reset."

func main() {
	kite.Event(event.DiscordMessageCreate, handleMessageCreateEvent)
}

func handleMessageCreateEvent(req event.Event) error {
	msg := req.Data.(distype.MessageCreateEvent)

	if count, err := strconv.Atoi(msg.Content); err == nil {
		if err := updateCounter(msg.ChannelID, count); err != nil {
			return err
		}
	}

	return nil
}

func updateCounter(channelID distype.Snowflake, count int) error {
	store := kv.New()

	counter, err := store.Increase(channelID.String(), 1)
	if err != nil {
		return err
	}

	if count != counter.Int() {
		if _, err := store.Delete(channelID.String()); err != nil {
			return err
		}

		if _, err := discord.MessageCreate(distype.MessageCreateRequest{
			ChannelID: channelID,
			MessageCreateParams: distype.MessageCreateParams{
				Content: &resetMessage,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
