package engine

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type InteractionData struct {
	interaction *discord.InteractionEvent
}

func (d *InteractionData) Interaction() *discord.InteractionEvent {
	return d.interaction
}

func (d *InteractionData) GuildID() discord.GuildID {
	return d.interaction.GuildID
}

func (d *InteractionData) ChannelID() discord.ChannelID {
	return d.interaction.ChannelID
}

func (d *InteractionData) CommandData() *discord.CommandInteraction {
	data, _ := d.interaction.Data.(*discord.CommandInteraction)
	return data
}

func (d *InteractionData) MessageComponentData() discord.ComponentInteraction {
	data, _ := d.interaction.Data.(discord.ComponentInteraction)
	return data
}

func (d *InteractionData) EventData() gateway.Event {
	return nil
}

type EventData struct{}
