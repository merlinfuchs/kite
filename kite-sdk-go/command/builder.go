package command

import (
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
)

func (c *Command) WithType(t distype.ApplicationCommandType) *Command {
	c.Type = t
	return c
}

func (c *Command) WithName(name string) *Command {
	c.Name = name
	return c
}

func (c *Command) WithDescription(description string) *Command {
	c.Description = description
	return c
}

func (c *Command) WithHandler(handler func(distype.Interaction, []distype.ApplicationCommandDataOption) error) *Command {
	c.Handler = handler
	return c
}

func (c *Command) WithDefaultMemberPermissions(permissions string) *Command {
	c.DefaultMemberPermissions = permissions
	return c
}

func (c *Command) WithDMPermission(dmPermission bool) *Command {
	c.DMPermission = &dmPermission
	return c
}

func (c *Command) WithNSFW(nsfw bool) *Command {
	c.NSFW = &nsfw
	return c
}

func (c *Command) WithOption(option CommandOption) *Command {
	c.Options = append(c.Options, manifest.DiscordCommandOption(option))
	return c
}

func (c *Command) WithOptions(options []CommandOption) *Command {
	res := make([]manifest.DiscordCommandOption, len(options))
	for i, opt := range options {
		res[i] = manifest.DiscordCommandOption(opt)
	}

	c.Options = res
	return c
}

func (o *CommandOption) WithChoice(choice CommandOptionChoice) *CommandOption {
	o.Choices = append(o.Choices, manifest.DiscordCommandOptionChoice(choice))
	return o
}
