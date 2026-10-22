package gateway

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store"
)

// maxDeletedMessages is how many deleted messages are kept per app. Flows read
// a deleted message right after the delete event, so a small buffer covers
// even bursts like a purge, without holding on to much memory per app.
const maxDeletedMessages = 250

type deletedMessageKey struct {
	channelID discord.ChannelID
	messageID discord.MessageID
}

// deletedMessageStore keeps messages after they are removed from the cache.
//
// Discord's MESSAGE_DELETE event only carries the IDs of the message, and the
// state removes the message from its cache before any handler sees the event.
// This store takes a copy of the cached message when it's removed, so Message
// Delete event listeners can still get its content and author. It's only
// known for messages the bot received while connected, and only for the most
// recent messages of each channel the state keeps.
type deletedMessageStore struct {
	store.MessageStore

	mu      sync.Mutex
	deleted map[deletedMessageKey]discord.Message
	// order is a ring of the kept keys. Once it's full, next is the oldest.
	order []deletedMessageKey
	next  int
}

func newDeletedMessageStore(inner store.MessageStore) *deletedMessageStore {
	return &deletedMessageStore{
		MessageStore: inner,
		deleted:      make(map[deletedMessageKey]discord.Message),
	}
}

// MessageRemove remembers the message before removing it from the cache. DMs
// are left out, as Message Delete listeners never run for them.
func (s *deletedMessageStore) MessageRemove(channelID discord.ChannelID, messageID discord.MessageID) error {
	if msg, err := s.MessageStore.Message(channelID, messageID); err == nil && msg != nil && msg.GuildID.IsValid() {
		s.remember(deletedMessageKey{channelID: channelID, messageID: messageID}, *msg)
	}
	return s.MessageStore.MessageRemove(channelID, messageID)
}

func (s *deletedMessageStore) remember(key deletedMessageKey, msg discord.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.deleted[key]; !ok {
		if len(s.order) < maxDeletedMessages {
			s.order = append(s.order, key)
		} else {
			delete(s.deleted, s.order[s.next])
			s.order[s.next] = key
			s.next = (s.next + 1) % maxDeletedMessages
		}
	}
	s.deleted[key] = msg
}

// DeletedMessage returns a message that was removed from the cache, if it's
// still known.
func (s *deletedMessageStore) DeletedMessage(channelID discord.ChannelID, messageID discord.MessageID) (discord.Message, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg, ok := s.deleted[deletedMessageKey{channelID: channelID, messageID: messageID}]
	return msg, ok
}

// trackDeletedMessages makes the cache of session keep deleted messages.
func trackDeletedMessages(session *state.State) {
	session.Cabinet.MessageStore = newDeletedMessageStore(session.Cabinet.MessageStore)
}
