package counting

import (
	"context"
	"fmt"
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

	enabled, err := c.GetValue(c, countEnabledKey(e.ChannelID.String()))
	if err != nil {
		return err
	}

	if !enabled.Bool() {
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
		c,
		countValueKey(e.ChannelID.String()),
		provider.VariableOperationIncrement,
		thing.New(1),
	)
	if err != nil {
		return err
	}

	lastUser, err := c.GetValue(c, countLastUserKey(e.ChannelID.String()))
	if err != nil {
		return err
	}

	if lastUser.String() == e.Author.ID.String() {
		err := cleanupCounting(c, e.ChannelID.String())
		if err != nil {
			return err
		}

		_, err = c.Discord().CreateMessage(c, e.ChannelID, api.SendMessageData{
			Content: fmt.Sprintf("%s RUINED IT AT **%d**. **You can't count two numbers in a row.**", e.Author.Mention(), nextNum.Int()-1),
			Reference: &discord.MessageReference{
				MessageID: e.ID,
			},
		})
		if err != nil {
			return err
		}

		return nil
	}

	if nextNum.Int() != num {
		err := cleanupCounting(c, e.ChannelID.String())
		if err != nil {
			return err
		}

		_, err = c.Discord().CreateMessage(c, e.ChannelID, api.SendMessageData{
			Content: fmt.Sprintf("%s RUINED IT AT **%d**. The next number was **%d**.", e.Author.Mention(), nextNum.Int()-1, nextNum.Int()),
			Reference: &discord.MessageReference{
				MessageID: e.ID,
			},
		})
		if err != nil {
			return err
		}

		return nil
	}

	_, err = c.UpdateValue(
		c,
		countLastUserKey(e.ChannelID.String()),
		provider.VariableOperationOverwrite,
		thing.New(e.Author.ID.String()),
	)
	if err != nil {
		return err
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
		err := c.DeleteValue(c, countEnabledKey(event.ChannelID.String()))
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
		_, err := c.UpdateValue(
			c,
			countEnabledKey(event.ChannelID.String()),
			provider.VariableOperationOverwrite,
			thing.New(true),
		)
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

func cleanupCounting(c plugin.Context, channelID string) error {
	err := c.DeleteValue(c, countValueKey(channelID))
	if err != nil {
		return err
	}

	err = c.DeleteValue(c, countLastUserKey(channelID))
	if err != nil {
		return err
	}

	return nil
}

func countEnabledKey(channelID string) string {
	return "counting-enabled:" + channelID
}

func countValueKey(channelID string) string {
	return "counting-value:" + channelID
}

func countLastUserKey(channelID string) string {
	return "counting-last-user:" + channelID
}
