package tickets

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

type SetupState string

const (
	SetupStateInit        SetupState = "init"
	SetupStateAdminRole   SetupState = "admin_role"
	SetupStateMenuChannel SetupState = "menu_channel"
	SetupStateLogChannel  SetupState = "log_channel"
	SetupStateComplete    SetupState = "complete"
)

const (
	SetupCustomIDInit         discord.ComponentID = "tickets_setup:init"
	SetupCustomIDAdminRole    discord.ComponentID = "tickets_setup:admin_role"
	SetupCustomIDMenuChannel  discord.ComponentID = "tickets_setup:menu_channel"
	SetupCustomIDLogChannel   discord.ComponentID = "tickets_setup:log_channel"
	PanelCustomIDCreateTicket discord.ComponentID = "tickets_menu:create_ticket"
)

func valueKeyAdminRole(guildID discord.GuildID) string {
	return fmt.Sprintf("admin_role:%d", guildID)
}

func valueKeyMenuChannel(guildID discord.GuildID) string {
	return fmt.Sprintf("menu_channel:%d", guildID)
}

func valueKeyLogChannel(guildID discord.GuildID) string {
	return fmt.Sprintf("log_channel:%d", guildID)
}

func getSetupResponseData(state SetupState) (api.InteractionResponse, error) {
	switch state {
	case SetupStateInit:
		return api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{
					{
						Title:       "👋 Welcome to Tickets!",
						Description: "Start the setup process with the button below. This process usually only takes a few minutes.",
						Color:       0x5765f2,
					},
				},
				Components: &discord.ContainerComponents{
					&discord.ActionRowComponent{
						&discord.ButtonComponent{
							Label:    "Start Setup",
							Style:    discord.PrimaryButtonStyle(),
							CustomID: SetupCustomIDInit,
						},
					},
				},
				Flags: discord.EphemeralMessage,
			},
		}, nil
	case SetupStateAdminRole:
		return api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{
					{
						Title:       "🎟️ Select Admin Role",
						Description: "Select the role that will be able to manage tickets.",
						Color:       0x5765f2,
					},
				},
				Components: &discord.ContainerComponents{
					&discord.ActionRowComponent{
						&discord.RoleSelectComponent{
							CustomID: SetupCustomIDAdminRole,
						},
					},
				},
				Flags: discord.EphemeralMessage,
			},
		}, nil
	case SetupStateMenuChannel:
		return api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{
					{
						Title:       "🎟️ Select Panel Channel",
						Description: "Select the channel that will be used by users to create tickets.",
						Color:       0x5765f2,
					},
				},
				Components: &discord.ContainerComponents{
					&discord.ActionRowComponent{
						&discord.ChannelSelectComponent{
							CustomID:     SetupCustomIDMenuChannel,
							ChannelTypes: []discord.ChannelType{discord.GuildText},
						},
					},
				},
				Flags: discord.EphemeralMessage,
			},
		}, nil
	case SetupStateLogChannel:
		return api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{
					{
						Title:       "🎟️ Select Log Channel",
						Description: "Select the channel that will be used for logging ticket events.",
						Color:       0x5765f2,
					},
				},
				Components: &discord.ContainerComponents{
					&discord.ActionRowComponent{
						&discord.ChannelSelectComponent{
							CustomID:     SetupCustomIDLogChannel,
							ChannelTypes: []discord.ChannelType{discord.GuildText},
						},
					},
				},
				Flags: discord.EphemeralMessage,
			},
		}, nil
	case SetupStateComplete:
		return api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{
					{
						Title:       "🎟️ Tickets Setup Complete",
						Description: "Tickets setup complete.",
						Color:       0x5765f2,
					},
				},
				Components: &discord.ContainerComponents{},
			},
		}, nil
	default:
		return api.InteractionResponse{}, fmt.Errorf("invalid menu state")
	}
}

func getPanelMessageData() api.SendMessageData {
	return api.SendMessageData{
		Content: "Hello, world!",
		Embeds: []discord.Embed{
			{
				Title:       "🎟️ Create Ticket",
				Description: "Create a new ticket.",
				Color:       0x5765f2,
			},
		},
		Components: discord.ContainerComponents{
			&discord.ActionRowComponent{
				&discord.ButtonComponent{
					Label:    "Create Ticket",
					Style:    discord.PrimaryButtonStyle(),
					CustomID: PanelCustomIDCreateTicket,
				},
			},
		},
	}
}
