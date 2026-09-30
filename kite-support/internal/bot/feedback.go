package bot

import (
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-support/internal/llm"
)

type forumThreadCreateData struct {
	Name                string                  `json:"name"`
	AutoArchiveDuration discord.ArchiveDuration `json:"auto_archive_duration,omitempty"`
	Message             api.SendMessageData     `json:"message"`
	AppliedTags         []discord.TagID         `json:"applied_tags,omitempty"`
}

func (b *Bot) createForumPost(channelID discord.ChannelID, data forumThreadCreateData) (*discord.Channel, error) {
	var ch *discord.Channel
	return ch, b.state.Client.RequestJSON(
		&ch, "POST",
		api.EndpointChannels+channelID.String()+"/threads",
		httputil.WithJSONBody(data),
	)
}

type feedbackContext struct {
	Question string
	Answer   string
	Intent   string
	// The questions and answers so far, this one included, for follow-ups.
	History   []llm.Turn
	UserID    discord.UserID
	ExpiresAt time.Time
}

type feedbackCache struct {
	mu   sync.Mutex
	data map[discord.MessageID]feedbackContext
}

func newFeedbackCache() *feedbackCache {
	return &feedbackCache{data: map[discord.MessageID]feedbackContext{}}
}

func (c *feedbackCache) Put(id discord.MessageID, ctx feedbackContext) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gcLocked()
	c.data[id] = ctx
}

func (c *feedbackCache) Get(id discord.MessageID) (feedbackContext, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gcLocked()
	ctx, ok := c.data[id]
	return ctx, ok
}

func (c *feedbackCache) Take(id discord.MessageID) (feedbackContext, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gcLocked()
	ctx, ok := c.data[id]
	if ok {
		delete(c.data, id)
	}
	return ctx, ok
}

func (c *feedbackCache) gcLocked() {
	now := time.Now()
	for id, ctx := range c.data {
		if now.After(ctx.ExpiresAt) {
			delete(c.data, id)
		}
	}
}
