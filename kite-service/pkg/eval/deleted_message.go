package eval

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

// DeletedMessageStore is a message store that keeps messages after they were
// deleted. The gateway sets one up for every app.
type DeletedMessageStore interface {
	DeletedMessage(channelID discord.ChannelID, messageID discord.MessageID) (discord.Message, bool)
}

// deletedMessage returns a deleted message from the cache of session, if the
// bot saw it before it was deleted.
func deletedMessage(session *state.State, channelID discord.ChannelID, messageID discord.MessageID) (discord.Message, bool) {
	if session == nil || session.Cabinet == nil {
		return discord.Message{}, false
	}

	store, ok := session.Cabinet.MessageStore.(DeletedMessageStore)
	if !ok {
		return discord.Message{}, false
	}
	return store.DeletedMessage(channelID, messageID)
}
