package engine

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/voicestate"
)

type InteractionData struct {
	interaction *discord.InteractionEvent
}

func (d *InteractionData) Interaction() *discord.InteractionEvent {
	return d.interaction
}

func (d *InteractionData) UserID() discord.UserID {
	return d.interaction.SenderID()
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

func (d *EventData) UserID() discord.UserID {
	switch data := d.event.(type) {
	case *gateway.MessageCreateEvent:
		return data.Author.ID
	case *gateway.MessageUpdateEvent:
		return data.Author.ID
	case *gateway.GuildMemberAddEvent:
		return data.User.ID
	case *gateway.GuildMemberRemoveEvent:
		return data.User.ID
	case *gateway.GuildMemberUpdateEvent:
		return data.User.ID
	case *voicestate.Event:
		return data.UserID
	}
	return 0
}

func (d *EventData) GuildID() discord.GuildID {
	switch data := d.event.(type) {
	case *gateway.MessageCreateEvent:
		return data.GuildID
	case *gateway.MessageDeleteEvent:
		return data.GuildID
	case *gateway.MessageUpdateEvent:
		return data.GuildID
	case *gateway.GuildMemberAddEvent:
		return data.GuildID
	case *gateway.GuildMemberRemoveEvent:
		return data.GuildID
	case *gateway.GuildMemberUpdateEvent:
		return data.GuildID
	case *state.GuildJoinEvent:
		return data.ID
	case *state.GuildLeaveEvent:
		return data.ID
	case *voicestate.Event:
		return data.GuildID
	}
	return 0
}

func (d *EventData) ChannelID() discord.ChannelID {
	switch data := d.event.(type) {
	case *gateway.MessageCreateEvent:
		return data.ChannelID
	case *gateway.MessageDeleteEvent:
		return data.ChannelID
	case *gateway.MessageUpdateEvent:
		return data.ChannelID
	case *voicestate.Event:
		// After leaving there's no current channel, so it's the one they left.
		if data.ChannelID.IsValid() {
			return data.ChannelID
		}
		return data.OldChannelID
	}
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
