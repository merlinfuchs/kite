package guildevent

import (
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

// PollEndType isn't a Discord event. Discord has none for polls ending, Kite
// derives it from the poll result message.
const PollEndType ws.EventType = "MESSAGE_POLL_END"

// PollResultMessage is the type of the message Discord posts when a poll
// ended and its votes are counted. arikawa has no constant for it.
const PollResultMessage discord.MessageType = 46

// pollResultEmbed is the type of the embed of the poll result message, which
// has the results as fields.
const pollResultEmbed discord.EmbedType = "poll_result"

// PollEnd is a poll that ended, from the poll result message Discord posts.
// The message is in the channel of the poll, by the author of the poll.
type PollEnd struct {
	// Message is the poll result message.
	Message discord.Message `json:"message"`
	// Member is the author of the poll as a member of the server.
	Member *discord.Member `json:"member,omitempty"`
	// PollMessageID is the message of the poll.
	PollMessageID discord.MessageID `json:"poll_message_id"`
	Result        PollResult        `json:"result"`
}

// PollResult is what the poll result message says about the poll.
type PollResult struct {
	Question   string `json:"question"`
	TotalVotes int    `json:"total_votes"`
	// WinnerVotes is the number of votes of the answers with the most votes.
	WinnerVotes int `json:"winner_votes"`
	// WinnerAnswerID is 0 if no answer has the most votes alone, a tie or a
	// poll without votes.
	WinnerAnswerID int    `json:"winner_answer_id,omitempty"`
	WinnerText     string `json:"winner_text,omitempty"`
	// WinnerEmoji is the emoji of the winning answer, nil without one.
	WinnerEmoji *discord.Emoji `json:"winner_emoji,omitempty"`
}

func (e *PollEnd) Op() ws.OpCode           { return 0 }
func (e *PollEnd) EventType() ws.EventType { return PollEndType }
func (e *PollEnd) derived()                {}

// NewPollEnd returns the poll that ended for a poll result message in a
// server, nil for other messages.
func NewPollEnd(e *gateway.MessageCreateEvent) *PollEnd {
	if e.Type != PollResultMessage || !e.GuildID.IsValid() {
		return nil
	}

	res := &PollEnd{Message: e.Message, Member: e.Member}
	if e.Reference != nil {
		res.PollMessageID = e.Reference.MessageID
	}
	for _, embed := range e.Embeds {
		if embed.Type == pollResultEmbed {
			res.Result = parsePollResult(embed.Fields)
			break
		}
	}
	return res
}

// parsePollResult reads the fields of the poll result embed. Discord sends
// them all as text.
func parsePollResult(fields []discord.EmbedField) PollResult {
	values := make(map[string]string, len(fields))
	for _, f := range fields {
		values[f.Name] = f.Value
	}

	number := func(name string) int {
		n, _ := strconv.Atoi(values[name])
		return n
	}

	res := PollResult{
		Question:       values["poll_question_text"],
		TotalVotes:     number("total_votes"),
		WinnerVotes:    number("victor_answer_votes"),
		WinnerAnswerID: number("victor_answer_id"),
		WinnerText:     values["victor_answer_text"],
	}

	if name := values["victor_answer_emoji_name"]; name != "" {
		emoji := &discord.Emoji{
			Name:     name,
			Animated: values["victor_answer_emoji_animated"] == "true",
		}
		if id, err := discord.ParseSnowflake(values["victor_answer_emoji_id"]); err == nil && id.IsValid() {
			emoji.ID = discord.EmojiID(id)
		}
		res.WinnerEmoji = emoji
	}
	return res
}
