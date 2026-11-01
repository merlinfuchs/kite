// Package pollevent has Discord's poll vote events, which arikawa doesn't know
// about. They're added to arikawa's gateway events, so the gateway decodes
// them and stored triggers find them like arikawa's own events.
package pollevent

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

const (
	VoteAddType    ws.EventType = "MESSAGE_POLL_VOTE_ADD"
	VoteRemoveType ws.EventType = "MESSAGE_POLL_VOTE_REMOVE"
)

// IntentGuildMessagePolls makes Discord send the votes on polls in servers.
// arikawa has no constant for it.
const IntentGuildMessagePolls gateway.Intents = 1 << 24

func init() {
	gateway.OpUnmarshalers.Add(
		func() ws.Event { return new(VoteAddEvent) },
		func() ws.Event { return new(VoteRemoveEvent) },
	)
}

// Vote is the data of both vote events. Discord only sends IDs.
type Vote struct {
	UserID    discord.UserID    `json:"user_id"`
	ChannelID discord.ChannelID `json:"channel_id"`
	// MessageID is the message of the poll.
	MessageID discord.MessageID `json:"message_id"`
	// GuildID is zero for polls in DMs.
	GuildID discord.GuildID `json:"guild_id,omitempty"`
	// AnswerID is the number of the answer, starting at 1 for the first.
	AnswerID int `json:"answer_id"`
}

// VoteAddEvent is MESSAGE_POLL_VOTE_ADD, a user voted for an answer. A user
// that votes for several answers of a poll that allows it sends one for each.
type VoteAddEvent struct {
	Vote
}

func (e *VoteAddEvent) Op() ws.OpCode           { return 0 }
func (e *VoteAddEvent) EventType() ws.EventType { return VoteAddType }

// VoteRemoveEvent is MESSAGE_POLL_VOTE_REMOVE, a user took back their vote
// for an answer.
type VoteRemoveEvent struct {
	Vote
}

func (e *VoteRemoveEvent) Op() ws.OpCode           { return 0 }
func (e *VoteRemoveEvent) EventType() ws.EventType { return VoteRemoveType }
