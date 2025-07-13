package counting

import (
	"context"
	"strconv"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type CountingPluginInstance struct {
	plugin *CountingPlugin
	appID  string
	config plugin.ConfigValues
}

func (p *CountingPluginInstance) Update(ctx context.Context, config plugin.ConfigValues) error {
	p.config = config
	return nil
}

func (p *CountingPluginInstance) HandleEvent(c plugin.Context, event gateway.Event) error {
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

	value, err := c.GetValue(c, e.ChannelID.String())
	if err != nil {
		return err
	}

	if value == thing.Null {
		return nil
	}

	if e.Content == "" {
		return nil
	}

	num, err := strconv.ParseInt(e.Content, 10, 64)
	if err != nil || num <= 0 {
		return nil
	}

	nextNum, err := c.UpdateValue(
		c, e.ChannelID.String(),
		provider.VariableOperationIncrement,
		thing.New(1),
	)
	if err != nil {
		return err
	}

	if nextNum.Int() != num {
		err := c.DeleteValue(c, e.ChannelID.String())
		if err != nil {
			return err
		}

		_, err = c.Discord().CreateMessage(c, e.ChannelID, api.SendMessageData{
			Content: "The count is incorrect. The next number is " + nextNum.String(),
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

func (p *CountingPluginInstance) HandleCommand(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	data, ok := event.Data.(*discord.CommandInteraction)
	if !ok {
		return nil
	}

	if data.Name != "counting-toggle" {
		return nil
	}

	value, err := c.GetValue(c, event.ChannelID.String())
	if err != nil {
		return err
	}

	if value != thing.Null {
		err := c.DeleteValue(c, event.ChannelID.String())
		if err != nil {
			return err
		}

		_, err = c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString("Counting is now disabled."),
				Flags:   discord.EphemeralMessage,
			},
		})
		if err != nil {
			return err
		}
	} else {
		_, err := c.UpdateValue(c, event.ChannelID.String(), provider.VariableOperationOverwrite, thing.New(0))
		if err != nil {
			return err
		}

		_, err = c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString("Counting is now enabled."),
				Flags:   discord.EphemeralMessage,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *CountingPluginInstance) HandleComponent(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *CountingPluginInstance) HandleModal(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *CountingPluginInstance) Close() error {
	return nil
}
