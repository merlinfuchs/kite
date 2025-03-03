package counting

import (
	"slices"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type CountingPluginInstance struct {
	config plugin.ConfigValues
}

func (p *CountingPluginInstance) Update(c plugin.Context) (plugin.UpdateResult, error) {
	return plugin.UpdateResult{}, nil
}

func (p *CountingPluginInstance) Events() []plugin.Event {
	return []plugin.Event{
		{
			ID:          "counting_message_create",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeMessageCreate,
			Description: "Check if the message is a counting message",
		},
	}
}

func (p *CountingPluginInstance) Commands() []plugin.Command {
	return []plugin.Command{
		{
			ID: "counting_toggle",
			Data: discord.Command{
				Name:        "toggle",
				Description: "Toggle the counting game in the current channel",
			},
		},
	}
}

func (p *CountingPluginInstance) HandleEvent(c plugin.Context, event gateway.Event) error {
	e, ok := event.(*gateway.MessageCreateEvent)
	if !ok {
		return nil
	}

	if e.Content == "!ping" {
		_, err := c.Client().SendMessage(e.ChannelID, "Pong!")
		if err != nil {
			return err
		}
	}

	channelIDs := p.config.GetStringArray(channelsConfigKey)
	if !slices.Contains(channelIDs, e.ChannelID.String()) {
		return nil
	}

	if e.Content == "" {
		return nil
	}

	num, err := strconv.Atoi(e.Content)
	if err != nil || num <= 0 {
		return nil
	}

	nextNum, err := c.IncreaseKey(e.ChannelID.String(), 1)
	if err != nil {
		return err
	}

	if nextNum != num {
		_, err := c.DeleteKey(e.ChannelID.String())
		if err != nil {
			return err
		}

		// TODO: Send message to channel that the count is incorrect

		return nil
	}

	// TODO: add checkmark reaction to the message

	return nil
}
