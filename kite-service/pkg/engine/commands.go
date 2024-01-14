package engine

import (
	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
)

func (e *PluginEngine) CommandsForGuild(guildID string) []*discordgo.ApplicationCommand {
	res := []*discordgo.ApplicationCommand{}

	plugins := append(e.StaticPlugins, e.Plugins...)

	for _, pl := range plugins {
		if _, ok := pl.GuildIDs[guildID]; ok {
			commands := convertCommands(pl.Plugin.Manifest().Commands)
			res = append(res, commands...)
		}
	}

	return res
}

func (e *PluginEngine) Commands() []*discordgo.ApplicationCommand {
	res := []*discordgo.ApplicationCommand{}

	plugins := append(e.StaticPlugins, e.Plugins...)

	for _, pl := range plugins {
		commands := convertCommands(pl.Plugin.Manifest().Commands)
		res = append(res, commands...)
	}

	return res
}

func convertCommands(commands []plugin.ManifestCommand) []*discordgo.ApplicationCommand {
	res := make([]*discordgo.ApplicationCommand, len(commands))

	for i, cmd := range commands {
		res[i] = &discordgo.ApplicationCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			// TODO: Options:     convertOptions(cmd.Options),
		}
	}

	return res
}
