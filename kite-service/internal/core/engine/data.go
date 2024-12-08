package engine

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/ws"
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

func (d *InteractionData) Event() ws.Event {
	return nil
}

type EventData struct {
	event ws.Event
}

func (d *EventData) Interaction() *discord.InteractionEvent {
	return nil
}

func (d *EventData) GuildID() discord.GuildID {
	return 0
}

func (d *EventData) ChannelID() discord.ChannelID {
	return 0
}

func (d *EventData) CommandData() *discord.CommandInteraction {
	return nil
}

func (d *EventData) MessageComponentData() discord.ComponentInteraction {
	return nil
}

func (d *EventData) Event() ws.Event {
	return d.event
}
