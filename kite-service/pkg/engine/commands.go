package engine

import (
	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
)

func (e *Engine) GuildCommands(guildID distype.Snowflake) []*discordgo.ApplicationCommand {
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
