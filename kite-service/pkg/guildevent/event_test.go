package guildevent

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func TestNewKnowsEveryEvent(t *testing.T) {
	for _, e := range []Event{&MessageDelete{}, &PollEnd{}} {
		got := New(e.EventType())
		if got == nil || got.EventType() != e.EventType() {
			t.Errorf("New(%q) = %T", e.EventType(), got)
		}
	}
	if New("GUILD_BAN_ADD") != nil {
		t.Error("New returned an event for a Discord event type")
	}
}

func TestNewPollEnd(t *testing.T) {
	e := &gateway.MessageCreateEvent{Message: discord.Message{
		Type:      PollResultMessage,
		GuildID:   1,
		ChannelID: 2,
		Reference: &discord.MessageReference{MessageID: 3, ChannelID: 2},
		Embeds: []discord.Embed{{
			Type: "poll_result",
			Fields: []discord.EmbedField{
				{Name: "poll_question_text", Value: "Pizza or pasta?"},
				{Name: "victor_answer_votes", Value: "4"},
				{Name: "total_votes", Value: "7"},
				{Name: "victor_answer_id", Value: "1"},
				{Name: "victor_answer_text", Value: "Pizza"},
				{Name: "victor_answer_emoji_name", Value: "🍕"},
			},
		}},
	}}

	got := NewPollEnd(e)
	if got == nil {
		t.Fatal("NewPollEnd returned nil for a poll result message")
	}
	if got.PollMessageID != 3 {
		t.Errorf("PollMessageID = %d, want 3", got.PollMessageID)
	}
	want := PollResult{Question: "Pizza or pasta?", TotalVotes: 7, WinnerVotes: 4, WinnerAnswerID: 1, WinnerText: "Pizza"}
	res := got.Result
	res.WinnerEmoji = nil
	if res != want {
		t.Errorf("Result = %+v, want %+v", res, want)
	}
	if got.Result.WinnerEmoji == nil || got.Result.WinnerEmoji.Name != "🍕" || got.Result.WinnerEmoji.ID.IsValid() {
		t.Errorf("WinnerEmoji = %+v, want the unicode emoji", got.Result.WinnerEmoji)
	}

	// A tie has no winning answer.
	e.Embeds[0].Fields = e.Embeds[0].Fields[:3]
	if got := NewPollEnd(e); got.Result.WinnerAnswerID != 0 || got.Result.WinnerText != "" || got.Result.WinnerEmoji != nil {
		t.Errorf("tie has a winner: %+v", got.Result)
	}

	e.Type = discord.DefaultMessage
	if NewPollEnd(e) != nil {
		t.Error("NewPollEnd returned a poll for a normal message")
	}
	e.Type = PollResultMessage
	e.GuildID = 0
	if NewPollEnd(e) != nil {
		t.Error("NewPollEnd returned a poll for a DM")
	}
}
