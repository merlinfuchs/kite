package flow

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type forumTestProvider struct {
	provider.MockDiscordProvider

	forum     discord.Channel
	channelID discord.ChannelID
	post      api.StartForumThreadData
}

func (p *forumTestProvider) Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error) {
	return &p.forum, nil
}

func (p *forumTestProvider) StartThreadInForum(ctx context.Context, channelID discord.ChannelID, data api.StartForumThreadData) (*discord.Channel, error) {
	p.channelID = channelID
	p.post = data
	return &discord.Channel{ID: 9, ParentID: channelID, Name: data.Name}, nil
}

func newForumTestProvider() *forumTestProvider {
	return &forumTestProvider{forum: discord.Channel{
		ID:   3,
		Type: discord.GuildForum,
		AvailableTags: []discord.Tag{
			{ID: 11, Name: "Bug Report"},
			{ID: 12, Name: "Question"},
		},
	}}
}

// executeForumPost runs the block below an entry, as the block resumes from a
// component when it's the entry.
func executeForumPost(t *testing.T, p *forumTestProvider, dataJSON string) (*FlowContext, error) {
	var data FlowNodeData
	require.NoError(t, json.Unmarshal([]byte(dataJSON), &data))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&blockTestContextData{},
		FlowProviders{
			Discord: p,
			Log:     &provider.MockLogProvider{},
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	node := &CompiledFlowNode{
		ID:   "0",
		Type: FlowNodeTypeEntryEvent,
		Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{
			{ID: "1", Type: FlowNodeTypeActionForumPostCreate, Data: data},
		}},
	}
	return c, node.Execute(c)
}

func TestForumPostCreate(t *testing.T) {
	p := newForumTestProvider()
	c, err := executeForumPost(t, p, `{
		"channel_target": "3",
		"channel_data": {"name": "Login broken", "applied_tags": "bug report, 12"},
		"message_data": {"content": "It doesn't work"},
		"audit_log_reason": "because"
	}`)
	require.NoError(t, err)

	assert.Equal(t, discord.ChannelID(3), p.channelID)
	assert.Equal(t, "Login broken", p.post.Name)
	assert.Equal(t, []discord.TagID{11, 12}, p.post.AppliedTags)
	assert.Equal(t, "It doesn't work", p.post.Message.Content)
	assert.Equal(t, api.AuditLogReason("because"), p.post.AuditLogReason)

	result := c.GetNodeResult("1")
	assert.Equal(t, thing.TypeDiscordChannel, result.Type)
	assert.Equal(t, discord.ChannelID(9), result.DiscordChannel().ID)
}

func TestForumPostCreateErrors(t *testing.T) {
	tests := []struct {
		name        string
		channelData string
		err         string
	}{
		{"no name", `{"name": "  "}`, "post name is required"},
		{"unknown tag", `{"name": "Hi", "applied_tags": "Bug Report, Feature"}`, `forum has no tag "Feature"`},
		{"too many tags", `{"name": "Hi", "applied_tags": "a, b, c, d, e, f"}`, "at most 5 tags"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newForumTestProvider()
			_, err := executeForumPost(t, p, `{
				"channel_target": "3",
				"channel_data": `+tt.channelData+`,
				"message_data": {"content": "Hello"}
			}`)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
			assert.Zero(t, p.channelID, "no post is created")
		})
	}

	t.Run("not a forum", func(t *testing.T) {
		p := newForumTestProvider()
		p.forum.Type = discord.GuildText
		_, err := executeForumPost(t, p, `{
			"channel_target": "3",
			"channel_data": {"name": "Hi", "applied_tags": "Question"},
			"message_data": {"content": "Hello"}
		}`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a forum or media channel")
	})

	t.Run("invalid channel", func(t *testing.T) {
		p := newForumTestProvider()
		_, err := executeForumPost(t, p, `{
			"channel_target": "abc",
			"channel_data": {"name": "Hi"},
			"message_data": {"content": "Hello"}
		}`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid forum channel")
	})
}
