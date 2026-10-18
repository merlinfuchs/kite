package flow

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dmTestChannelID discord.ChannelID = 77

type dmTestProvider struct {
	provider.MockDiscordProvider

	dmErr          error
	privateChanErr error
	sentChannels   []discord.ChannelID
}

func (p *dmTestProvider) CreatePrivateChannel(ctx context.Context, userID discord.UserID) (*discord.Channel, error) {
	if p.privateChanErr != nil {
		return nil, p.privateChanErr
	}
	return &discord.Channel{ID: dmTestChannelID, Type: discord.DirectMessage}, nil
}

func (p *dmTestProvider) CreateMessage(ctx context.Context, channelID discord.ChannelID, data api.SendMessageData) (*discord.Message, error) {
	if channelID == dmTestChannelID && p.dmErr != nil {
		return nil, p.dmErr
	}
	p.sentChannels = append(p.sentChannels, channelID)
	return &discord.Message{ID: 1, ChannelID: channelID, Content: data.Content}, nil
}

type dmTestLogProvider struct {
	entries []string
}

func (p *dmTestLogProvider) CreateLogEntry(ctx context.Context, level provider.LogLevel, msg string) {
	p.entries = append(p.entries, string(level)+": "+msg)
}

// executeDirectMessage runs a Send direct message block below an entry, with
// a Create message block after it to show whether the flow carried on.
func executeDirectMessage(t *testing.T, p *dmTestProvider, logs *dmTestLogProvider) (*FlowContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&blockTestContextData{},
		FlowProviders{
			Discord: p,
			Log:     logs,
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	after := &CompiledFlowNode{
		ID:   "2",
		Type: FlowNodeTypeActionMessageCreate,
		Data: FlowNodeData{
			ChannelTarget: "5",
			MessageData:   &message.MessageData{Content: "after"},
		},
	}
	dm := &CompiledFlowNode{
		ID:   "1",
		Type: FlowNodeTypeActionPrivateMessageCreate,
		Data: FlowNodeData{
			UserTarget:  "42",
			MessageData: &message.MessageData{Content: "hi"},
		},
		Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{after}},
	}
	node := &CompiledFlowNode{
		ID:       "0",
		Type:     FlowNodeTypeEntryEvent,
		Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{dm}},
	}
	return c, node.Execute(c)
}

func TestDirectMessageSent(t *testing.T) {
	p := &dmTestProvider{}
	c, err := executeDirectMessage(t, p, &dmTestLogProvider{})
	require.NoError(t, err)

	assert.Equal(t, []discord.ChannelID{dmTestChannelID, 5}, p.sentChannels)
	assert.False(t, c.GetNodeResult("1").IsNil())
}

func TestDirectMessageClosedContinuesFlow(t *testing.T) {
	closed := &httputil.HTTPError{Status: 403, Code: discordErrorCannotSendToUser, Message: "Cannot send messages to this user"}
	noMutualGuilds := &httputil.HTTPError{Status: 403, Code: discordErrorNoMutualGuilds, Message: "Cannot send messages to this user due to having no mutual guilds"}

	for name, p := range map[string]*dmTestProvider{
		"message rejected":        {dmErr: closed},
		"private channel refused": {privateChanErr: closed},
		// The engine's provider wraps Discord errors like this.
		"no mutual guilds": {dmErr: fmt.Errorf("failed to send message: %w", noMutualGuilds)},
	} {
		t.Run(name, func(t *testing.T) {
			logs := &dmTestLogProvider{}
			c, err := executeDirectMessage(t, p, logs)
			require.NoError(t, err)

			// The block after the DM still ran.
			assert.Equal(t, []discord.ChannelID{5}, p.sentChannels)
			assert.True(t, c.GetNodeResult("1").IsNil())
			require.Len(t, logs.entries, 1)
			assert.Contains(t, logs.entries[0], "warn: ")
		})
	}
}

func TestDirectMessageOtherErrorsStillFail(t *testing.T) {
	p := &dmTestProvider{dmErr: errors.New("boom")}
	_, err := executeDirectMessage(t, p, &dmTestLogProvider{})
	require.Error(t, err)
	assert.Empty(t, p.sentChannels)
}
