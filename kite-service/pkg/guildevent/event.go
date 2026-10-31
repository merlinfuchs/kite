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

// MessageDeleteType is the type of Discord's own event, as MessageDelete
// replaces that event for event listeners.
const MessageDeleteType ws.EventType = "MESSAGE_DELETE"

// MemberUpdateType is the type of Discord's own event, as MemberUpdate
// replaces that event for event listeners.
const MemberUpdateType ws.EventType = "GUILD_MEMBER_UPDATE"

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

// MessageEdit is Discord's MESSAGE_UPDATE when the message was edited. Discord
// also sends MESSAGE_UPDATE when it adds a link preview or the message is
// pinned, which must not run Message Update listeners again. It has the type
// of Discord's event.
type MessageEdit struct {
	gateway.MessageUpdateEvent
}

func (e *MessageEdit) derived() {}

// MemberUpdate is Discord's GUILD_MEMBER_UPDATE with the member from before
// the update. Discord sends the whole member, also when only its roles,
// timeout or flags changed, so event listeners get this instead of Discord's
// event, and only for the changes Member Update listeners are for. It has the
// same fields as Discord's event, so stored ones of those decode into it.
type MemberUpdate struct {
	gateway.GuildMemberUpdateEvent
	// Old is the member as it was cached before the update.
	Old discord.Member `json:"old"`
}

func (e *MemberUpdate) Op() ws.OpCode           { return 0 }
func (e *MemberUpdate) EventType() ws.EventType { return MemberUpdateType }
func (e *MemberUpdate) derived()                {}

// NicknameChanged reports whether the nickname in the server was set, changed
// or removed.
func (e *MemberUpdate) NicknameChanged() bool {
	return e.Old.Nick != e.Nick
}

// AvatarChanged reports whether the server avatar was set, changed or
// removed. The avatar of the account itself comes with USER_UPDATE instead.
func (e *MemberUpdate) AvatarChanged() bool {
	return e.Old.Avatar != e.Avatar
}

// ScreeningPassed reports whether the member just accepted the rules of the
// server's membership screening.
func (e *MemberUpdate) ScreeningPassed() bool {
	return e.Old.IsPending && !e.IsPending
}

// Changes lists what changed, in the order of the event filter values.
func (e *MemberUpdate) Changes() []string {
	var res []string
	if e.NicknameChanged() {
		res = append(res, MemberChangeNickname)
	}
	if e.AvatarChanged() {
		res = append(res, MemberChangeAvatar)
	}
	if e.ScreeningPassed() {
		res = append(res, MemberChangeScreeningPassed)
	}
	return res
}

// The changes MemberUpdate is derived for, as the Event Filter and the
// placeholders name them.
const (
	MemberChangeNickname        = "nickname"
	MemberChangeAvatar          = "avatar"
	MemberChangeScreeningPassed = "screening_passed"
)

// New returns an empty event of a type of this package, nil for other types.
// Used to decode stored events. Discord's MESSAGE_DELETE and
// GUILD_MEMBER_UPDATE never run a flow, so stored ones are always a
// MessageDelete or MemberUpdate.
func New(t ws.EventType) Event {
	switch t {
	case MessageDeleteType:
		return new(MessageDelete)
	case MemberUpdateType:
		return new(MemberUpdate)
	}
	return nil
}
