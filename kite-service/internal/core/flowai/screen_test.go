package flowai

import (
	"context"
	"errors"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDecisions struct {
	noul float64
	err  error
	opts provider.DecideOpts
}

func (d *fakeDecisions) Decide(ctx context.Context, opts provider.DecideOpts) (provider.Decision, error) {
	d.opts = opts
	return provider.Decision{Noul: d.noul}, d.err
}

var nukeRequest = Request{
	Flow: "entry_command({ name: \"nuke\" }, [])",
	Messages: []Message{
		{Role: "user", Content: "make a nuke command"},
		{Role: "assistant", Content: "Done."},
		{Role: "user", Content: "fix the errors"},
	},
	Issues: []string{"'Create channel' never runs"},
}

func TestRespondRefuses(t *testing.T) {
	assistant, body := fakeOpenAI(t, "completed", `{"message": "Done.", "edits": []}`)
	d := &fakeDecisions{noul: 0.95}
	assistant.config.Decisions = d

	_, err := assistant.Respond(context.Background(), nukeRequest)
	assert.ErrorIs(t, err, ErrRefused)
	assert.Nil(t, *body, "OpenAI was called")
	assert.Equal(t, provider.DecisionTypeNoul, d.opts.Type)
	assert.Equal(t, "Request: make a nuke command\n\nRequest: fix the errors\n\nProblems to fix:\n- 'Create channel' never runs\n\nCurrent flow:\nentry_command({ name: \"nuke\" }, [])", d.opts.Input)
}

func TestRespondAnswersBelowTheThreshold(t *testing.T) {
	assistant, body := fakeOpenAI(t, "completed", `{"message": "Done.", "edits": []}`)
	assistant.config.Decisions = &fakeDecisions{noul: 0.6}

	_, err := assistant.Respond(context.Background(), nukeRequest)
	require.NoError(t, err)
	assert.NotNil(t, *body)
}

func TestRespondAnswersIfTheCheckFails(t *testing.T) {
	assistant, body := fakeOpenAI(t, "completed", `{"message": "Done.", "edits": []}`)
	assistant.config.Decisions = &fakeDecisions{noul: 0.95, err: errors.New("unavailable")}

	_, err := assistant.Respond(context.Background(), nukeRequest)
	require.NoError(t, err)
	assert.NotNil(t, *body)
}
