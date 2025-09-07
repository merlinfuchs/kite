package tickets

import (
	"context"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type TicketsPluginInstance struct {
	appID  string
	config plugin.ConfigValues
}

func (p *TicketsPluginInstance) Update(ctx context.Context, config plugin.ConfigValues) error {
	p.config = config
	return nil
}

func (p *TicketsPluginInstance) HandleEvent(c plugin.Context, event gateway.Event) error {
	return nil
}

func (p *TicketsPluginInstance) HandleCommand(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	data, ok := event.Data.(*discord.CommandInteraction)
	if !ok {
		return nil
	}

	if data.Name != "tickets" || len(data.Options) == 0 {
		return nil
	}

	subCMD := data.Options[0]

	switch subCMD.Name {
	case "setup":
		resp, err := getSetupResponseData(SetupStateInit)
		if err != nil {
			return err
		}

		_, err = c.Discord().CreateInteractionResponse(c, event.ID, event.Token, resp)
		if err != nil {
			return err
		}

		return nil
	case "disable":
		return nil
	}

	return nil
}

func (p *TicketsPluginInstance) HandleComponent(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	data, ok := event.Data.(discord.ComponentInteraction)
	if !ok {
		return nil
	}

	var err error
	var resp api.InteractionResponse

	switch data.ID() {
	case SetupCustomIDInit:
		resp, err = getSetupResponseData(SetupStateAdminRole)
	case SetupCustomIDAdminRole:
		// TODO: Save admin role to config
		comp, ok := data.(*discord.RoleSelectInteraction)
		if !ok {
			return nil
		}
		if len(comp.Values) == 0 {
			return nil
		}

		_, err = c.UpdateValue(
			c,
			valueKeyAdminRole(event.GuildID),
			provider.VariableOperationOverwrite,
			thing.NewString(comp.Values[0].String()),
			provider.WithMetadata("guild_id", event.GuildID.String()),
		)
		if err != nil {
			return err
		}

		resp, err = getSetupResponseData(SetupStateMenuChannel)
	case SetupCustomIDMenuChannel:
		comp, ok := data.(*discord.ChannelSelectInteraction)
		if !ok {
			return nil
		}
		if len(comp.Values) == 0 {
			return nil
		}

		_, err = c.Discord().CreateMessage(c, comp.Values[0], getPanelMessageData())
		if err != nil {
			return fmt.Errorf("failed to create menu message: %w", err)
		}

		_, err = c.UpdateValue(
			c,
			valueKeyMenuChannel(event.GuildID),
			provider.VariableOperationOverwrite,
			thing.NewString(comp.Values[0].String()),
			provider.WithMetadata("guild_id", event.GuildID.String()),
		)
		if err != nil {
			return err
		}

		resp, err = getSetupResponseData(SetupStateLogChannel)
	case SetupCustomIDLogChannel:
		comp, ok := data.(*discord.ChannelSelectInteraction)
		if !ok {
			return nil
		}
		if len(comp.Values) == 0 {
			return nil
		}

		_, err = c.UpdateValue(
			c,
			valueKeyLogChannel(event.GuildID),
			provider.VariableOperationOverwrite,
			thing.NewString(comp.Values[0].String()),
			provider.WithMetadata("guild_id", event.GuildID.String()),
		)
		if err != nil {
			return err
		}

		resp, err = getSetupResponseData(SetupStateComplete)
	case PanelCustomIDCreateTicket:
		resp = api.InteractionResponse{
			Type: api.ModalResponse,
			Data: &api.InteractionResponseData{
				CustomID: option.NewNullableString(string(PanelCustomIDCreateTicket)),
				Title:    option.NewNullableString("Create Ticket"),
				Components: &discord.ContainerComponents{
					&discord.ActionRowComponent{
						&discord.TextInputComponent{
							CustomID:    "topic",
							Label:       "Topic",
							Style:       discord.TextInputParagraphStyle,
							Required:    true,
							Placeholder: "Enter the topic of the ticket",
							Value:       "",
						},
					},
				},
			},
		}
	}

	if err != nil {
		return err
	}

	_, err = c.Discord().CreateInteractionResponse(c, event.ID, event.Token, resp)
	if err != nil {
		return err
	}

	return nil
}

func (p *TicketsPluginInstance) HandleModal(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	data, ok := event.Data.(*discord.ModalInteraction)
	if !ok {
		return nil
	}

	if data.CustomID != PanelCustomIDCreateTicket {
		return nil
	}
	return nil
}

func (p *TicketsPluginInstance) Close() error {
	return nil
}
