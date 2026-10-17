package tickets

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

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
		Description: "Let members open private support tickets with a button.",
		Icon:        "ticket",
		Author:      "Kite",
	}
}

func (p *TicketsPlugin) Config() plugin.Config {
	return plugin.Config{}
}

func (p *TicketsPlugin) Events() []plugin.Event {
	return []plugin.Event{}
}

func (p *TicketsPlugin) Commands() []plugin.Command {
	setupPerms := discord.PermissionManageGuild

	return []plugin.Command{
		{
			ID: "cmd_tickets",
			Data: api.CreateCommandData{
				Name:                     "tickets",
				Description:              "Set up the ticket system for the current server",
				DefaultMemberPermissions: &setupPerms,
				Contexts:                 guildOnly,
				Options: discord.CommandOptions{
					&discord.SubcommandOption{
						OptionName:  "setup",
						Description: "Post the ticket panel and save the ticket settings",
						Options: []discord.CommandOptionValue{
							&discord.ChannelOption{
								OptionName:   "channel",
								Description:  "The channel to post the panel with the Open Ticket button in",
								Required:     true,
								ChannelTypes: []discord.ChannelType{discord.GuildText, discord.GuildAnnouncement},
							},
							&discord.ChannelOption{
								OptionName:   "category",
								Description:  "The category to create ticket channels in",
								Required:     false,
								ChannelTypes: []discord.ChannelType{discord.GuildCategory},
							},
							&discord.RoleOption{
								OptionName:  "support-role",
								Description: "The role that can see and answer all tickets",
								Required:    false,
							},
							&discord.ChannelOption{
								OptionName:   "log-channel",
								Description:  "The channel to log opened and closed tickets in",
								Required:     false,
								ChannelTypes: []discord.ChannelType{discord.GuildText, discord.GuildAnnouncement},
							},
							&discord.StringOption{
								OptionName:  "message",
								Description: "The text of the panel",
								Required:    false,
								MaxLength:   optionInt(1000),
							},
						},
					},
					&discord.SubcommandOption{
						OptionName:  "disable",
						Description: "Stop new tickets from being opened. Open tickets stay.",
					},
				},
			},
		},
		{
			ID: "cmd_ticket",
			Data: api.CreateCommandData{
				Name:        "ticket",
				Description: "Manage the current ticket",
				Contexts:    guildOnly,
				Options: discord.CommandOptions{
					&discord.SubcommandOption{
						OptionName:  "close",
						Description: "Close the current ticket",
					},
					&discord.SubcommandOption{
						OptionName:  "add",
						Description: "Add a member to the current ticket",
						Options: []discord.CommandOptionValue{
							&discord.UserOption{
								OptionName:  "user",
								Description: "The member to add",
								Required:    true,
							},
						},
					},
					&discord.SubcommandOption{
						OptionName:  "remove",
						Description: "Remove a member from the current ticket",
						Options: []discord.CommandOptionValue{
							&discord.UserOption{
								OptionName:  "user",
								Description: "The member to remove",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}

// guildOnly hides the commands in DMs, tickets only exist inside a server.
var guildOnly = []discord.InteractionContext{discord.InteractionContextGuild}

func optionInt(v int) option.Int {
	return &v
}
