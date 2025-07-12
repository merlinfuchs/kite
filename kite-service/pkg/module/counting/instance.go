package counting

import (
	"slices"
	"strconv"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/module"
)

type CountingModuleInstance struct {
	appID  string
	config module.ConfigValues
}

func (p *CountingModuleInstance) Update(c module.Context, config module.ConfigValues) error {
	p.config = config
	return nil
}

func (p *CountingModuleInstance) Events() []module.Event {
	return []module.Event{
		{
			ID:          "counting_message_create",
			Source:      module.EventSourceDiscord,
			Type:        module.EventTypeMessageCreate,
			Description: "Check if the message is a counting message",
		},
	}
}

func (p *CountingModuleInstance) Commands() []module.Command {
	return []module.Command{
		{
			ID: "counting_toggle",
			Data: api.CreateCommandData{
				Name:        "toggle",
				Description: "Toggle the counting game in the current channel",
			},
		},
	}
}

func (p *CountingModuleInstance) HandleEvent(c module.Context, event gateway.Event) error {
	e, ok := event.(*gateway.MessageCreateEvent)
	if !ok {
		return nil
	}

	if e.Content == "!ping" {
		_, err := c.Discord().CreateMessage(c, e.ChannelID, api.SendMessageData{
			Content: "Pong!",
		})
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

		_, err = c.Discord().CreateMessage(c, e.ChannelID, api.SendMessageData{
			Content: "The count is incorrect. The next number is " + strconv.Itoa(nextNum),
		})
		if err != nil {
			return err
		}

		return nil
	}

	err = c.Discord().CreateMessageReaction(c, e.ChannelID, e.ID, "✅")
	if err != nil {
		return err
	}

	return nil
}
