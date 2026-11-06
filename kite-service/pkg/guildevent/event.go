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
const VoiceStateUpdateType ws.EventType = "VOICE_STATE_UPDATE"

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

type VoiceStateUpdate struct {
	gateway.VoiceStateUpdateEvent
	OldChannelID discord.ChannelID `json:"old_channel_id,omitempty"`
	Action       string            `json:"action"`
}

func (e *VoiceStateUpdate) Op() ws.OpCode           { return 0 }
func (e *VoiceStateUpdate) EventType() ws.EventType { return VoiceStateUpdateType }
func (e *VoiceStateUpdate) derived()                {}

// New returns an empty event of a type of this package, nil for other types.
// Used to decode stored events. Discord's MESSAGE_DELETE never runs a flow, so
// a stored one is always a MessageDelete.
func New(t ws.EventType) Event {
	switch t {
	case MessageDeleteType:
		return new(MessageDelete)
	case VoiceStateUpdateType:
		return new(VoiceStateUpdate)
	}
	return nil
}
