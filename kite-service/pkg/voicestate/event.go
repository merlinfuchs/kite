// Package voicestate defines the event that triggers the flows of voice state
// event listeners.
package voicestate

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

const EventType ws.EventType = "VOICE_STATE_UPDATE"

type Action string

const (
	ActionJoin  Action = "joined"
	ActionLeave Action = "left"
	ActionMove  Action = "moved"
)

// Event is Discord's VOICE_STATE_UPDATE together with the channel the user was
// in before it. Discord only sends the new voice state, so the old channel has
// to be read from the state cache before the update is applied to it.
type Event struct {
	discord.VoiceState

	// OldChannelID is unset if the user wasn't in a voice channel before, or
	// their voice state wasn't cached.
	OldChannelID discord.ChannelID `json:"old_channel_id"`
}

// Op is the dispatch op code, like every other event that triggers flows.
func (e *Event) Op() ws.OpCode { return 0 }

func (e *Event) EventType() ws.EventType { return EventType }

// Action tells what the user did. It's empty if they stayed in the same
// channel, which is the case when they mute, deafen or start streaming.
func (e *Event) Action() Action {
	switch {
	case !e.ChannelID.IsValid():
		// Leaving without a cached state still is a leave, there's just no
		// channel to report.
		return ActionLeave
	case !e.OldChannelID.IsValid():
		return ActionJoin
	case e.OldChannelID != e.ChannelID:
		return ActionMove
	}
	return ""
}
