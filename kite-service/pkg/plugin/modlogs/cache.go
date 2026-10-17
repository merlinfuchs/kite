package modlogs

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
)

// messageCacheSize is how many recent messages each app remembers. Discord
// doesn't send the content of deleted messages or the content from before an
// edit, so the plugin can only log what it saw arrive.
const messageCacheSize = 5000

type cachedMessage struct {
	GuildID     discord.GuildID
	ChannelID   discord.ChannelID
	Author      discord.User
	Content     string
	Attachments int
}

// messageCache keeps the most recent messages up to a fixed size, dropping the
// oldest one when a new one doesn't fit anymore.
type messageCache struct {
	mu      sync.Mutex
	entries map[discord.MessageID]cachedMessage
	ring    []discord.MessageID
	next    int
}

func newMessageCache(size int) *messageCache {
	return &messageCache{
		entries: make(map[discord.MessageID]cachedMessage, size),
		ring:    make([]discord.MessageID, size),
	}
}

// Set adds a message or updates it if it's still cached. Updating a message
// doesn't make it stay longer.
func (c *messageCache) Set(id discord.MessageID, msg cachedMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.entries[id]; ok {
		c.entries[id] = msg
		return
	}

	if len(c.ring) == 0 {
		return
	}

	delete(c.entries, c.ring[c.next])
	c.ring[c.next] = id
	c.next = (c.next + 1) % len(c.ring)
	c.entries[id] = msg
}

func (c *messageCache) Get(id discord.MessageID) (cachedMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg, ok := c.entries[id]
	return msg, ok
}

// Delete removes a message and returns it if it was cached.
func (c *messageCache) Delete(id discord.MessageID) (cachedMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg, ok := c.entries[id]
	delete(c.entries, id)
	return msg, ok
}
