package engine

import (
	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/go-types/manifest"
)

func (e *PluginEngine) GuildCommands(guildID string) []*discordgo.ApplicationCommand {
	res := []*discordgo.ApplicationCommand{}

	for _, pl := range e.Deployments[guildID] {
		commands := convertCommands(pl.Manifest().DiscordCommands)
		res = append(res, commands...)
	}

	return res
}

func convertCommands(commands []manifest.DiscordCommand) []*discordgo.ApplicationCommand {
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
