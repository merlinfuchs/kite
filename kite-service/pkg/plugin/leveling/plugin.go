package leveling

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type LevelingPlugin struct {
}

func NewLevelingPlugin() *LevelingPlugin {
	return &LevelingPlugin{}
}

func (p *LevelingPlugin) Instance(ctx context.Context, appID string, config plugin.ConfigValues) (plugin.PluginInstance, error) {
	return &LevelingPluginInstance{
		appID:     appID,
		config:    config,
		cooldowns: make(map[string]int64),
	}, nil
}

func (p *LevelingPlugin) ID() string {
	return "leveling"
}

func (p *LevelingPlugin) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "Leveling",
		Description: "Give members XP for chatting, with levels, a leaderboard and role rewards.",
		Icon:        "trophy",
		Author:      "Kite",
	}
}

func (p *LevelingPlugin) Config() plugin.Config {
	return plugin.Config{}
}

func (p *LevelingPlugin) Events() []plugin.Event {
	return []plugin.Event{
		{
			ID:          "event_message_create",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeMessageCreate,
			Description: "Give the author of the message XP and announce when they reach a new level",
		},
	}
}

func (p *LevelingPlugin) Commands() []plugin.Command {
	adminPerms := discord.PermissionManageGuild

	return []plugin.Command{
		{
			ID: "cmd_rank",
			Data: api.CreateCommandData{
				Name:        "rank",
				Description: "Show the level and XP of a member",
				Contexts:    guildOnly,
				Options: discord.CommandOptions{
					&discord.UserOption{
						OptionName:  "user",
						Description: "The member to show, yourself if empty",
						Required:    false,
					},
				},
			},
		},
		{
			ID: "cmd_leaderboard",
			Data: api.CreateCommandData{
				Name:        "leaderboard",
				Description: "Show the members with the most XP",
				Contexts:    guildOnly,
			},
		},
		{
			ID: "cmd_leveling",
			Data: api.CreateCommandData{
				Name:                     "leveling",
				Description:              "Configure leveling for the current server",
				DefaultMemberPermissions: &adminPerms,
				Contexts:                 guildOnly,
				Options: discord.CommandOptions{
					&discord.SubcommandOption{
						OptionName:  "enable",
						Description: "Start giving members XP for their messages",
						Options: []discord.CommandOptionValue{
							&discord.ChannelOption{
								OptionName:   "channel",
								Description:  "The channel to announce new levels in, the channel of the message if empty",
								Required:     false,
								ChannelTypes: []discord.ChannelType{discord.GuildText, discord.GuildAnnouncement},
							},
							&discord.StringOption{
								OptionName:  "message",
								Description: "The level up message, {user} and {level} are replaced",
								Required:    false,
								MaxLength:   optionInt(500),
							},
						},
					},
					&discord.SubcommandOption{
						OptionName:  "disable",
						Description: "Stop giving XP. Members keep the XP they have.",
					},
					&discord.SubcommandOption{
						OptionName:  "reward-add",
						Description: "Give a role to members when they reach a level",
						Options: []discord.CommandOptionValue{
							&discord.IntegerOption{
								OptionName:  "level",
								Description: "The level that gives the role",
								Required:    true,
								Min:         optionInt(1),
								Max:         optionInt(maxLevel),
							},
							&discord.RoleOption{
								OptionName:  "role",
								Description: "The role to give",
								Required:    true,
							},
						},
					},
					&discord.SubcommandOption{
						OptionName:  "reward-remove",
						Description: "Stop giving a role for a level",
						Options: []discord.CommandOptionValue{
							&discord.IntegerOption{
								OptionName:  "level",
								Description: "The level of the reward",
								Required:    true,
								Min:         optionInt(1),
								Max:         optionInt(maxLevel),
							},
						},
					},
					&discord.SubcommandOption{
						OptionName:  "reset",
						Description: "Reset the XP of a member to zero",
						Options: []discord.CommandOptionValue{
							&discord.UserOption{
								OptionName:  "user",
								Description: "The member to reset",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}

var guildOnly = []discord.InteractionContext{discord.InteractionContextGuild}

func optionInt(v int) option.Int {
	return &v
}
