package eval

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

// PollEnv is the poll of a vote, or of a poll that ended. The message of the
// event is the poll.
type PollEnv struct {
	// AnswerID is 0 for a poll that ended.
	AnswerID int `expr:"answer_id" json:"answer_id"`

	// The results are only known when the poll ended.
	TotalVotes  int    `expr:"total_votes" json:"total_votes"`
	WinnerID    int    `expr:"winner_id" json:"winner_id"`
	Winner      string `expr:"winner" json:"winner"`
	WinnerEmoji string `expr:"winner_emoji" json:"winner_emoji"`
	WinnerVotes int    `expr:"winner_votes" json:"winner_votes"`

	// poll returns the question and answers of the poll. Votes only have the
	// IDs, so the poll is only fetched when an expression reads them.
	poll func() (*pollData, error)
}

// pollData is the part of a poll the lazy fields read.
type pollData struct {
	Question struct {
		Text string `json:"text"`
	} `json:"question"`
	Answers []struct {
		AnswerID  int `json:"answer_id"`
		PollMedia struct {
			Text  string         `json:"text"`
			Emoji *discord.Emoji `json:"emoji"`
		} `json:"poll_media"`
	} `json:"answers"`
}

// NewPollVoteEnv is the poll of a vote for the answer answerID. fetch is only
// called once, when an expression first reads the question or the answer.
func NewPollVoteEnv(answerID int, fetch func() (*pollData, error)) *PollEnv {
	return &PollEnv{
		AnswerID: answerID,
		poll:     sync.OnceValues(fetch),
	}
}

// NewPollEndEnv is a poll that ended, from its result message, which has the
// question, so nothing needs to be fetched.
func NewPollEndEnv(r guildevent.PollResult) *PollEnv {
	res := &PollEnv{
		TotalVotes:  r.TotalVotes,
		WinnerID:    r.WinnerAnswerID,
		Winner:      r.WinnerText,
		WinnerVotes: r.WinnerVotes,
	}
	if r.WinnerEmoji != nil {
		res.WinnerEmoji = r.WinnerEmoji.String()
	}

	data := &pollData{}
	data.Question.Text = r.Question
	res.poll = func() (*pollData, error) { return data, nil }
	return res
}

// fetchPoll gets the poll of a message from Discord. arikawa's message has no
// poll, so the request is built by hand.
func fetchPoll(session *state.State, channelID discord.ChannelID, messageID discord.MessageID) (*pollData, error) {
	if session == nil {
		return &pollData{}, nil
	}

	var msg struct {
		Poll *pollData `json:"poll"`
	}
	err := session.RequestJSON(
		&msg,
		http.MethodGet,
		api.EndpointChannels+channelID.String()+"/messages/"+messageID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get poll: %w", err)
	}
	if msg.Poll == nil {
		return &pollData{}, nil
	}
	return msg.Poll, nil
}

type pollLazyFields struct {
	Question    string `expr:"question" desc:"The question of the poll"`
	Answer      string `expr:"answer" desc:"The text of the answer voted for, empty when a poll ended"`
	AnswerEmoji string `expr:"answer_emoji" desc:"The emoji of the answer voted for, empty without one"`
}

func (p PollEnv) lazyFields() any {
	return pollLazyFields{}
}

// lazyField has the question and the answer voted for, which votes don't
// include.
func (p PollEnv) lazyField(name string) (any, bool, error) {
	switch name {
	case "question", "answer", "answer_emoji":
	default:
		return nil, false, nil
	}

	if p.poll == nil {
		return "", true, nil
	}
	data, err := p.poll()
	if err != nil {
		return "", true, err
	}

	if name == "question" {
		return data.Question.Text, true, nil
	}
	for _, a := range data.Answers {
		if a.AnswerID != p.AnswerID || p.AnswerID == 0 {
			continue
		}
		if name == "answer" {
			return a.PollMedia.Text, true, nil
		}
		if a.PollMedia.Emoji != nil {
			return a.PollMedia.Emoji.String(), true, nil
		}
	}
	return "", true, nil
}

// String is the answer voted for, or the winning answer of a poll that ended.
// A poll that ended without a winner is the question.
func (p PollEnv) String() string {
	if p.AnswerID != 0 {
		if answer, _, err := p.lazyField("answer"); err == nil && answer != "" {
			return answer.(string)
		}
		return strconv.Itoa(p.AnswerID)
	}
	if p.Winner != "" {
		return p.Winner
	}
	question, _, _ := p.lazyField("question")
	return question.(string)
}
