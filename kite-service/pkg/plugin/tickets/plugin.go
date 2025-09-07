package tickets

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

// Configuration: admin role, creation menu channel, log channel

// Features: Menu for creating tickets, Claim by admin, Close by admin or user, Transcripts?

// Events: Ticket created, Ticket claimed, Ticket closed

type TicketsPlugin struct {
}

func NewTicketsPlugin() *TicketsPlugin {
	return &TicketsPlugin{}
}

func (p *TicketsPlugin) Instance(ctx context.Context, appID string, config plugin.ConfigValues) (plugin.PluginInstance, error) {
	return &TicketsPluginInstance{
		appID:  appID,
		config: config,
	}, nil
}

func (p *TicketsPlugin) ID() string {
	return "tickets"
}

func (p *TicketsPlugin) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "Tickets",
		Description: "Add a ticket system to your app, so users can create tickets to ask questions, contact support, etc.",
		Icon:        "ticket",
		Author:      "Merlin",
	}
}

func (p *TicketsPlugin) Config() plugin.Config {
	return plugin.Config{}
}

func (p *TicketsPlugin) Events() []plugin.Event {
	return []plugin.Event{
		{
			ID:          "event_message_reaction_add",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeMessageReactionAdd,
			Description: "Check if the message has surpassed the threshold for the starboard",
		},
	}
}

func (p *TicketsPlugin) Commands() []plugin.Command {
	perms := discord.PermissionAdministrator

	return []plugin.Command{
		{
			ID: "cmd_ticket",
			Data: api.CreateCommandData{
				Name:                     "ticket",
				Description:              "Configure the ticket for the current server",
				DefaultMemberPermissions: &perms,
				Options: discord.CommandOptions{
					&discord.SubcommandOption{
						OptionName:  "enable",
						Description: "Enable the starboard for the current server",
						Options: []discord.CommandOptionValue{
							&discord.ChannelOption{
								OptionName:  "channel",
								Description: "The channel to use as the starboard",
								Required:    true,
							},
							&discord.IntegerOption{
								OptionName:  "threshold",
								Description: "The number of stars required to pin a message to the starboard",
								Required:    true,
							},
							&discord.StringOption{
								OptionName:  "emoji",
								Description: "The emoji to use instead of a ⭐️",
								Required:    false,
							},
						},
					},
					&discord.SubcommandOption{
						OptionName:  "disable",
						Description: "Disable the starboard for the current server",
					},
				},
			},
		},
	}
}
