package command

import (
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
)

type Command struct {
	manifest.DiscordCommand
	Handler func(distype.Interaction, []distype.ApplicationCommandDataOption) error
}

func New(name string) *Command {
	return &Command{
		DiscordCommand: manifest.DiscordCommand{
			Name: name,
		},
	}
}

func (c *Command) FullName() string {
	return c.Name
}

type CommandOption manifest.DiscordCommandOption

type CommandOptionChoice manifest.DiscordCommandOptionChoice
