// Package guildevent has the events Kite derives from Discord's server events,
// for things Discord's events leave out, like what a deleted message
// contained. They're worked out from the cache from before Discord's event
// updated it.
//
// They're gateway events, so they go through the same execution path as the
// Discord events they come from. MessageEdit only decides which updates run
// listeners, which get Discord's event inside it.
package guildevent

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

// MessageDeleteType and MemberRemoveType are the types of Discord's own
// events, as MessageDelete and MemberRemove replace those events for event
// listeners.
const (
	MessageDeleteType ws.EventType = "MESSAGE_DELETE"
	MemberRemoveType  ws.EventType = "GUILD_MEMBER_REMOVE"
)

// Event is implemented by every event of this package.
type Event interface {
	ws.Event
	derived()
}

// MessageDelete is Discord's MESSAGE_DELETE with the message that was deleted.
// Discord only sends the IDs, so event listeners get it instead of Discord's
// event. It has the same fields as Discord's event, so stored ones of those
// decode into it.
type MessageDelete struct {
	gateway.MessageDeleteEvent
	// Message is the deleted message as it was cached, nil if the bot didn't
	// see it.
	Message *discord.Message `json:"message,omitempty"`
}

func (e *MessageDelete) Op() ws.OpCode           { return 0 }
func (e *MessageDelete) EventType() ws.EventType { return MessageDeleteType }
func (e *MessageDelete) derived()                {}

// MemberRemove is Discord's GUILD_MEMBER_REMOVE with the member who left.
// Discord only sends the user and the server ID, so event listeners get it
// instead of Discord's event. It has the same fields as Discord's event, so
// stored ones of those decode into it.
type MemberRemove struct {
	gateway.GuildMemberRemoveEvent
	// Member is the member as they were cached before leaving, nil if the bot
	// hadn't seen them in the server.
	Member *discord.Member `json:"member,omitempty"`
}

func (e *MemberRemove) Op() ws.OpCode           { return 0 }
func (e *MemberRemove) EventType() ws.EventType { return MemberRemoveType }
func (e *MemberRemove) derived()                {}

// MessageEdit is Discord's MESSAGE_UPDATE when the message was edited. Discord
// also sends MESSAGE_UPDATE when it adds a link preview or the message is
// pinned, which must not run Message Update listeners again. It has the type
// of Discord's event.
type MessageEdit struct {
	gateway.MessageUpdateEvent
}

func (e *MessageEdit) derived() {}

// New returns an empty event of a type of this package, nil for other types.
// Used to decode stored events. Discord's MESSAGE_DELETE and
// GUILD_MEMBER_REMOVE never run a flow, so stored ones are always the derived
// events.
func New(t ws.EventType) Event {
	switch t {
	case MessageDeleteType:
		return new(MessageDelete)
	case MemberRemoveType:
		return new(MemberRemove)
	}
	return nil
}
