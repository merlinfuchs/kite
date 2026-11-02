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

// PresenceUpdateType is the type of Discord's own event, as PresenceChange
// replaces that event for event listeners.
const PresenceUpdateType ws.EventType = "PRESENCE_UPDATE"

// PresenceChange is Discord's PRESENCE_UPDATE when the status or the custom
// status of a member changed. Discord also sends PRESENCE_UPDATE for every
// change of a game, song or other activity, which must not run Presence Update
// listeners and spend credits each time. It has the same fields as Discord's
// event, so stored ones of those decode into it.
type PresenceChange struct {
	gateway.PresenceUpdateEvent
	// Previous is the presence from before the change as it was cached, nil
	// if the bot didn't know it. Discord only sends members who are online
	// when the bot connects, so a member who isn't cached was offline.
	Previous *discord.Presence `json:"previous,omitempty"`
}

func (e *PresenceChange) Op() ws.OpCode           { return 0 }
func (e *PresenceChange) EventType() ws.EventType { return PresenceUpdateType }
func (e *PresenceChange) derived()                {}

// CustomStatus returns the text and emoji of the custom status of a presence,
// empty if there is none. An offline member has no activities, so no custom
// status either.
func CustomStatus(p discord.Presence) (text string, emoji string) {
	for _, a := range p.Activities {
		if a.Type != discord.CustomActivity {
			continue
		}
		if a.Emoji != nil {
			// The mention of a custom emoji, or the unicode emoji.
			emoji = a.Emoji.String()
		}
		return a.State, emoji
	}
	return "", ""
}

// New returns an empty event of a type of this package, nil for other types.
// Used to decode stored events. Discord's MESSAGE_DELETE and PRESENCE_UPDATE never
// run a flow, so a stored one is always a MessageDelete or PresenceChange.
func New(t ws.EventType) Event {
	switch t {
	case MessageDeleteType:
		return new(MessageDelete)
	case PresenceUpdateType:
		return new(PresenceChange)
	}
	return nil
}
