package eval

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
	"github.com/kitecloud/kite/kite-service/pkg/pollevent"
)

func pollContext(env *EventEnv) Context {
	return NewContext(Env{
		"event":   env,
		"user":    env.User,
		"channel": env.Channel,
		"message": env.Message,
		"poll":    env.Poll,
	})
}

func testPoll(t *testing.T) *pollData {
	var data pollData
	err := json.Unmarshal([]byte(`{
		"question": {"text": "Pizza or pasta?"},
		"answers": [
			{"answer_id": 1, "poll_media": {"text": "Pizza", "emoji": {"id": null, "name": "🍕"}}},
			{"answer_id": 2, "poll_media": {"text": "Pasta"}}
		]
	}`), &data)
	if err != nil {
		t.Fatal(err)
	}
	return &data
}

func TestPollVotePlaceholders(t *testing.T) {
	vote := pollevent.Vote{UserID: 7, ChannelID: 3, MessageID: 4, GuildID: testGuildID, AnswerID: 1}
	env := NewEventEnv(&pollevent.VoteAddEvent{Vote: vote}, nil)

	fetches := 0
	data := testPoll(t)
	env.Poll = NewPollVoteEnv(1, func() (*pollData, error) {
		fetches++
		return data, nil
	})
	ctx := pollContext(env)

	if got := evalString(t, ctx, "{{user.id}}|{{message.id}}|{{poll.answer_id}}"); got != "7|4|1" {
		t.Errorf("got %q", got)
	}
	if fetches != 0 {
		t.Errorf("fetched the poll %d times for placeholders that don't need it", fetches)
	}

	got := evalString(t, ctx, "{{poll}}|{{poll.question}}|{{poll.answer}}|{{poll.answer_emoji}}")
	if want := "Pizza|Pizza or pasta?|Pizza|🍕"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if fetches != 1 {
		t.Errorf("fetched the poll %d times, want once", fetches)
	}

	// Pasta has no emoji.
	env.Poll = NewPollVoteEnv(2, func() (*pollData, error) { return data, nil })
	if got := evalString(t, pollContext(env), "{{poll.answer}}|{{poll.answer_emoji}}"); got != "Pasta|" {
		t.Errorf("got %q", got)
	}

	// Without a session nothing is fetched and the answer is empty.
	env = NewEventEnv(&pollevent.VoteRemoveEvent{Vote: vote}, nil)
	if got := evalString(t, pollContext(env), "{{poll.answer_id}}|{{poll.answer}}"); got != "1|" {
		t.Errorf("got %q", got)
	}
}

func TestPollVoteFetchError(t *testing.T) {
	env := NewEventEnv(&pollevent.VoteAddEvent{Vote: pollevent.Vote{AnswerID: 1}}, nil)
	env.Poll = NewPollVoteEnv(1, func() (*pollData, error) {
		return nil, errors.New("missing access")
	})
	if _, err := EvalTemplate(t.Context(), "{{poll.answer}}", pollContext(env)); err == nil {
		t.Error("a poll that couldn't be fetched didn't fail")
	}
}

func TestPollEndPlaceholders(t *testing.T) {
	end := &guildevent.PollEnd{
		Message: discord.Message{
			ID: 5, ChannelID: 3, GuildID: testGuildID,
			Author: discord.User{ID: 7, Username: "creator"},
		},
		PollMessageID: 4,
		Result: guildevent.PollResult{
			Question:       "Pizza or pasta?",
			TotalVotes:     7,
			WinnerVotes:    4,
			WinnerAnswerID: 1,
			WinnerText:     "Pizza",
			WinnerEmoji:    &discord.Emoji{Name: "🍕"},
		},
	}
	ctx := pollContext(NewEventEnv(end, nil))

	got := evalString(t, ctx, "{{user.id}}|{{message.id}}|{{poll}}|{{poll.question}}|{{poll.winner_id}}|{{poll.winner_emoji}}|{{poll.winner_votes}}/{{poll.total_votes}}|{{poll.answer}}")
	if want := "7|4|Pizza|Pizza or pasta?|1|🍕|4/7|"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// A tie has no winner.
	end.Result.WinnerAnswerID, end.Result.WinnerText, end.Result.WinnerEmoji = 0, "", nil
	ctx = pollContext(NewEventEnv(end, nil))
	if got := evalString(t, ctx, "{{poll}}|{{poll.winner}}|{{poll.winner_id}}"); got != "Pizza or pasta?||0" {
		t.Errorf("got %q", got)
	}
}
