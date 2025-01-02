package eval

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

type Env map[string]any

type InteractionEnv struct {
	interaction *discord.InteractionEvent

	ID        string      `expr:"id"`
	ChannelID string      `expr:"channel_id"`
	GuildID   string      `expr:"guild_id"`
	Command   *CommandEnv `expr:"command"`
}

func NewInteractionEnv(i *discord.InteractionEvent) *InteractionEnv {
	e := &InteractionEnv{
		interaction: i,

		ID:        i.ID.String(),
		ChannelID: i.ChannelID.String(),
		GuildID:   i.GuildID.String(),
	}

	if i.Data.InteractionType() == discord.CommandInteractionType {
		e.Command = NewCommandEnv(i)
	}

	return e
}

func NewEnvWithInteraction(i *discord.InteractionEvent) Env {
	interactionEnv := NewInteractionEnv(i)
	env := Env{
		"interaction": interactionEnv,
	}
	if interactionEnv.Command != nil {
		env["arg"] = interactionEnv.Command.ArgFunc
	}

	return env
}

func (e InteractionEnv) String() string {
	return e.interaction.ID.String()
}

type CommandEnv struct {
	interaction *discord.InteractionEvent
	cmd         *discord.CommandInteraction

	ArgFunc func(name string) *CommandOptionEnv `expr:"arg"`
}

func NewCommandEnv(i *discord.InteractionEvent) *CommandEnv {
	data, _ := i.Data.(*discord.CommandInteraction)

	return &CommandEnv{
		interaction: i,
		cmd:         data,

		ArgFunc: func(name string) *CommandOptionEnv {
			for _, option := range data.Options {
				if option.Name == name {
					return NewCommandOptionEnv(i, data, &option)
				}
			}

			return nil
		},
	}
}

func (c CommandEnv) String() string {
	return c.cmd.ID.String()
}

type CommandOptionEnv struct {
	interaction *discord.InteractionEvent
	cmd         *discord.CommandInteraction
	option      *discord.CommandInteractionOption
}

func NewCommandOptionEnv(
	i *discord.InteractionEvent,
	cmd *discord.CommandInteraction,
	option *discord.CommandInteractionOption,
) *CommandOptionEnv {
	return &CommandOptionEnv{
		interaction: i,
		cmd:         cmd,
		option:      option,
	}
}

func (o CommandOptionEnv) String() string {
	return o.option.String()
}

type EventEnv struct {
	event ws.Event

	User    *UserEnv    `expr:"user"`
	Member  *MemberEnv  `expr:"member"`
	Channel *ChannelEnv `expr:"channel"`
	Message *MessageEnv `expr:"message"`
	Guild   *GuildEnv   `expr:"guild"`
}

func NewEventEnv(event ws.Event) *EventEnv {
	return &EventEnv{
		event: event,
	}
}

func NewEnvWithEvent(event ws.Event) Env {
	return Env{
		"event": NewEventEnv(event),
	}
}

type UserEnv struct{}

type MemberEnv struct{}

type ChannelEnv struct{}

type MessageEnv struct{}

type GuildEnv struct{}
