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
	// order holds the keys oldest first, to drop the oldest message when
	// the store is full.
	order []deletedMessageKey
}

func newDeletedMessageStore(inner store.MessageStore) *deletedMessageStore {
	return &deletedMessageStore{
		MessageStore: inner,
		deleted:      make(map[deletedMessageKey]discord.Message),
	}
}

// MessageRemove remembers the message before removing it from the cache.
func (s *deletedMessageStore) MessageRemove(channelID discord.ChannelID, messageID discord.MessageID) error {
	if msg, err := s.MessageStore.Message(channelID, messageID); err == nil && msg != nil {
		s.remember(deletedMessageKey{channelID: channelID, messageID: messageID}, *msg)
	}
	return s.MessageStore.MessageRemove(channelID, messageID)
}

func (s *deletedMessageStore) remember(key deletedMessageKey, msg discord.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.deleted[key]; !ok {
		if len(s.order) >= maxDeletedMessages {
			delete(s.deleted, s.order[0])
			// Shift instead of reslicing, so the backing array doesn't grow
			// forever.
			copy(s.order, s.order[1:])
			s.order = s.order[:len(s.order)-1]
		}
		s.order = append(s.order, key)
	}
	s.deleted[key] = msg
}

// DeletedMessage returns a copy of a message that was removed from the cache,
// if it's still known.
func (s *deletedMessageStore) DeletedMessage(channelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg, ok := s.deleted[deletedMessageKey{channelID: channelID, messageID: messageID}]
	if !ok {
		return nil, false
	}
	return &msg, true
}

// trackDeletedMessages makes the cache of session keep deleted messages, which
// eval reads to fill in the placeholders of Message Delete events.
func trackDeletedMessages(session *state.State) {
	session.Cabinet.MessageStore = newDeletedMessageStore(session.Cabinet.MessageStore)
}
