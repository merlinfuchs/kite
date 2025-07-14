package starboard

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type StarboardPlugin struct {
}

func NewStarboardPlugin() *StarboardPlugin {
	return &StarboardPlugin{}
}

func (p *StarboardPlugin) Instance(ctx context.Context, appID string, config plugin.ConfigValues) (plugin.PluginInstance, error) {
	return &StarboardPluginInstance{
		appID:  appID,
		config: config,
	}, nil
}

func (p *StarboardPlugin) ID() string {
	return "starboard"
}

func (p *StarboardPlugin) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "Starboard",
		Description: "Create a starboard for your server.",
		Icon:        "star",
		Author:      "Merlin",
	}
}

func (p *StarboardPlugin) Config() plugin.Config {
	return plugin.Config{}
}

func (p *StarboardPlugin) Events() []plugin.Event {
	return []plugin.Event{
		{
			ID:          "event_message_reaction_add",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeMessageReactionAdd,
			Description: "Check if the message has surpassed the threshold for the starboard",
		},
	}
}

func (p *StarboardPlugin) Commands() []plugin.Command {
	perms := discord.PermissionAdministrator

	return []plugin.Command{
		{
			ID: "cmd_starboard",
			Data: api.CreateCommandData{
				Name:                     "starboard",
				Description:              "Configure the starboard for the current server",
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
