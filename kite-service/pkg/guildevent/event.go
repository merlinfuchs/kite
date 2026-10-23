// Package guildevent has the events Kite derives from Discord's server events,
// for things Discord has no event of its own for, like a role being added to a
// member. They're worked out by comparing Discord's event with the cache from
// before it.
//
// They're gateway events, so they go through the same execution path as the
// Discord events they come from.
package guildevent

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

const (
	MemberRoleAddType       ws.EventType = "GUILD_MEMBER_ROLE_ADD"
	MemberRoleRemoveType    ws.EventType = "GUILD_MEMBER_ROLE_REMOVE"
	MemberTimeoutAddType    ws.EventType = "GUILD_MEMBER_TIMEOUT_ADD"
	MemberTimeoutRemoveType ws.EventType = "GUILD_MEMBER_TIMEOUT_REMOVE"
	// ChannelUpdateType and MessageDeleteType are the types of Discord's own
	// events, as they replace those events for event listeners.
	ChannelUpdateType ws.EventType = "CHANNEL_UPDATE"
	MessageDeleteType ws.EventType = "MESSAGE_DELETE"
)

// Event is implemented by every event of this package.
type Event interface {
	ws.Event
	derived()
}

// MemberRole is a role added to or removed from a member. A change of several
// roles at once is one event per role.
type MemberRole struct {
	GuildID discord.GuildID `json:"guild_id"`
	// Member is the member after the change.
	Member discord.Member `json:"member"`
	RoleID discord.RoleID `json:"role_id"`
}

type MemberRoleAdd struct{ MemberRole }

func (e *MemberRoleAdd) Op() ws.OpCode           { return 0 }
func (e *MemberRoleAdd) EventType() ws.EventType { return MemberRoleAddType }
func (e *MemberRoleAdd) derived()                {}

type MemberRoleRemove struct{ MemberRole }

func (e *MemberRoleRemove) Op() ws.OpCode           { return 0 }
func (e *MemberRoleRemove) EventType() ws.EventType { return MemberRoleRemoveType }
func (e *MemberRoleRemove) derived()                {}

// MemberTimeout is a member being timed out, their timeout being changed to
// end at another time, or it being removed. Discord sends nothing when a
// timeout runs out.
type MemberTimeout struct {
	GuildID discord.GuildID `json:"guild_id"`
	// Member is the member after the change.
	Member discord.Member `json:"member"`
}

type MemberTimeoutAdd struct{ MemberTimeout }

func (e *MemberTimeoutAdd) Op() ws.OpCode           { return 0 }
func (e *MemberTimeoutAdd) EventType() ws.EventType { return MemberTimeoutAddType }
func (e *MemberTimeoutAdd) derived()                {}

type MemberTimeoutRemove struct{ MemberTimeout }

func (e *MemberTimeoutRemove) Op() ws.OpCode           { return 0 }
func (e *MemberTimeoutRemove) EventType() ws.EventType { return MemberTimeoutRemoveType }
func (e *MemberTimeoutRemove) derived()                {}

// ChannelUpdate is Discord's CHANNEL_UPDATE with the channel from before.
// Event listeners get it instead of Discord's event, so they can show what
// changed and skip updates that only move the channel.
type ChannelUpdate struct {
	Channel discord.Channel `json:"channel"`
	// Before is the channel before the update, nil if it wasn't known.
	Before *discord.Channel `json:"before,omitempty"`
	// Changes names what the update changed, empty if Before isn't known.
	Changes []string `json:"changes,omitempty"`
}

func (e *ChannelUpdate) Op() ws.OpCode           { return 0 }
func (e *ChannelUpdate) EventType() ws.EventType { return ChannelUpdateType }
func (e *ChannelUpdate) derived()                {}

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

// New returns an empty event of a type of this package, nil for other types.
// Used to decode stored events. Discord's CHANNEL_UPDATE and MESSAGE_DELETE
// never run a flow, so a stored one is always a ChannelUpdate or MessageDelete.
func New(t ws.EventType) Event {
	switch t {
	case MemberRoleAddType:
		return new(MemberRoleAdd)
	case MemberRoleRemoveType:
		return new(MemberRoleRemove)
	case MemberTimeoutAddType:
		return new(MemberTimeoutAdd)
	case MemberTimeoutRemoveType:
		return new(MemberTimeoutRemove)
	case ChannelUpdateType:
		return new(ChannelUpdate)
	case MessageDeleteType:
		return new(MessageDelete)
	}
	return nil
}
