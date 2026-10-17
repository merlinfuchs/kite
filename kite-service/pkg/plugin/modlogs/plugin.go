package modlogs

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type ModlogsPlugin struct {
}

func NewModlogsPlugin() *ModlogsPlugin {
	return &ModlogsPlugin{}
}

func (p *ModlogsPlugin) Instance(ctx context.Context, appID string, config plugin.ConfigValues) (plugin.PluginInstance, error) {
	return &ModlogsPluginInstance{
		appID:    appID,
		config:   config,
		messages: newMessageCache(messageCacheSize),
	}, nil
}

func (p *ModlogsPlugin) ID() string {
	return "modlogs"
}

func (p *ModlogsPlugin) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "Mod Logs",
		Description: "Log deleted and edited messages, joins, leaves, kicks, bans and timeouts to a channel.",
		Icon:        "scroll-text",
		Author:      "Kite",
	}
}

func (p *ModlogsPlugin) Config() plugin.Config {
	return plugin.Config{}
}

func (p *ModlogsPlugin) Events() []plugin.Event {
	return []plugin.Event{
		{
			ID:          "event_message_create",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeMessageCreate,
			Description: "Remember recent messages, so their content can be logged when they are edited or deleted",
		},
		{
			ID:          "event_message_update",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeMessageUpdate,
			Description: "Log edited messages",
		},
		{
			ID:          "event_message_delete",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeMessageDelete,
			Description: "Log deleted messages",
		},
		{
			ID:          "event_guild_member_add",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeGuildMemberAdd,
			Description: "Log members joining the server",
		},
		{
			ID:          "event_guild_member_remove",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeGuildMemberRemove,
			Description: "Log members leaving the server or being kicked",
		},
		{
			ID:          "event_guild_member_update",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeGuildMemberUpdate,
			Description: "Log members being timed out and their timeouts being removed",
		},
		{
			ID:          "event_guild_ban_add",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeGuildBanAdd,
			Description: "Log members being banned",
		},
		{
			ID:          "event_guild_ban_remove",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeGuildBanRemove,
			Description: "Log users being unbanned",
		},
	}
}

func (p *ModlogsPlugin) Commands() []plugin.Command {
	perms := discord.PermissionManageGuild

	return []plugin.Command{
		{
			ID: "cmd_modlogs",
			Data: api.CreateCommandData{
				Name:                     "modlogs",
				Description:              "Configure the mod logs for the current server",
				DefaultMemberPermissions: &perms,
				Contexts:                 []discord.InteractionContext{discord.InteractionContextGuild},
				Options: discord.CommandOptions{
					&discord.SubcommandOption{
						OptionName:  "enable",
						Description: "Start logging to a channel",
						Options: []discord.CommandOptionValue{
							&discord.ChannelOption{
								OptionName:   "channel",
								Description:  "The channel to send the logs to",
								Required:     true,
								ChannelTypes: []discord.ChannelType{discord.GuildText, discord.GuildAnnouncement},
							},
						},
					},
					&discord.SubcommandOption{
						OptionName:  "disable",
						Description: "Stop logging for this server",
					},
				},
			},
		},
	}
}
