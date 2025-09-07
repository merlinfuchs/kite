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
	return []plugin.Event{}
}

func (p *TicketsPlugin) Commands() []plugin.Command {
	perms := discord.PermissionAdministrator

	return []plugin.Command{
		{
			ID: "cmd_tickets",
			Data: api.CreateCommandData{
				Name:                     "tickets",
				Description:              "Configure the ticket system for the current server",
				DefaultMemberPermissions: &perms,
				Options: discord.CommandOptions{
					&discord.SubcommandOption{
						OptionName:  "setup",
						Description: "Setup the ticket system for the current server",
						Options:     []discord.CommandOptionValue{},
					},
					&discord.SubcommandOption{
						OptionName:  "disable",
						Description: "Disable the ticket system for the current server",
					},
				},
			},
		},
	}
}
