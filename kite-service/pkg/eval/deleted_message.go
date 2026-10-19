package eval

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

// DeletedMessageStore is a message store that keeps messages after they were
// deleted. The gateway sets one up for every app, so Message Delete events can
// provide the content and author of the message, which Discord doesn't send.
type DeletedMessageStore interface {
	DeletedMessage(channelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, bool)
}

// deletedMessage returns a deleted message from the cache of session, if the
// bot saw it before it was deleted.
func deletedMessage(session *state.State, channelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, bool) {
	if session == nil || session.Cabinet == nil {
		return nil, false
	}

	store, ok := session.Cabinet.MessageStore.(DeletedMessageStore)
	if !ok {
		return nil, false
	}
	return store.DeletedMessage(channelID, messageID)
}

// cachedMember returns a copy of a member from the cache of session.
func cachedMember(session *state.State, guildID discord.GuildID, userID discord.UserID) (*discord.Member, bool) {
	if session == nil || session.Cabinet == nil || guildID == 0 || userID == 0 {
		return nil, false
	}

	member, err := session.Cabinet.Member(guildID, userID)
	if err != nil || member == nil {
		return nil, false
	}
	cpy := *member
	return &cpy, true
}
