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

func TestRefuses(t *testing.T) {
	req := Request{
		Flow: "entry_command({ name: \"nuke\" }, [])",
		Messages: []Message{
			{Role: "user", Content: "make a nuke command"},
			{Role: "assistant", Content: "Done."},
			{Role: "user", Content: "fix the errors"},
		},
	}

	d := &fakeDecisions{noul: 0.95}
	refused, err := NewAssistant(nil, Config{Decisions: d}).Refuses(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, refused)
	assert.Equal(t, provider.DecisionTypeNoul, d.opts.Type)
	assert.Equal(t, "Request: make a nuke command\n\nRequest: fix the errors\n\nCurrent flow:\nentry_command({ name: \"nuke\" }, [])", d.opts.Input)

	d.noul = 0.6
	refused, err = NewAssistant(nil, Config{Decisions: d}).Refuses(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, refused)

	d.err = errors.New("unavailable")
	_, err = NewAssistant(nil, Config{Decisions: d}).Refuses(context.Background(), req)
	assert.Error(t, err)

	refused, err = NewAssistant(nil, Config{}).Refuses(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, refused)
}
