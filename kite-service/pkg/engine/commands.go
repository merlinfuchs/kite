package engine

import (
	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
)

func (e *PluginEngine) GuildCommands(guildID string) []*discordgo.ApplicationCommand {
	res := []*discordgo.ApplicationCommand{}

	for _, pl := range e.Deployments[guildID] {
		commands := convertCommands(pl.Manifest().Commands)
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
