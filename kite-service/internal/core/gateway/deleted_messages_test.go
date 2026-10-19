package gateway

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
)

func TestTrackDeletedMessages(t *testing.T) {
	session := state.New("Bot token")
	trackDeletedMessages(session)

	deleted, ok := session.Cabinet.MessageStore.(eval.DeletedMessageStore)
	if !ok {
		t.Fatal("message store doesn't keep deleted messages")
	}

	// Events go through the session like they do when Discord sends them, so
	// the state removes the message from its cache before handlers run.
	dispatch := func(e any) { session.Session.Handler.Call(e) }

	const channelID = discord.ChannelID(10)
	author := discord.User{ID: 5, Username: "jack"}

	dispatch(&gateway.MessageCreateEvent{Message: discord.Message{
		ID:        1000,
		ChannelID: channelID,
		GuildID:   20,
		Author:    author,
		Content:   "hello",
	}})
	dispatch(&gateway.MessageUpdateEvent{Message: discord.Message{
		ID:        1000,
		ChannelID: channelID,
		GuildID:   20,
		Author:    author,
		Content:   "hello, edited",
	}})

	if _, ok := deleted.DeletedMessage(channelID, 1000); ok {
		t.Fatal("message is known as deleted before it was deleted")
	}

	dispatch(&gateway.MessageDeleteEvent{ID: 1000, ChannelID: channelID, GuildID: 20})

	if _, err := session.Cabinet.Message(channelID, 1000); err == nil {
		t.Fatal("deleted message is still in the cache")
	}

	msg, ok := deleted.DeletedMessage(channelID, 1000)
	if !ok {
		t.Fatal("deleted message wasn't kept")
	}
	if msg.Content != "hello, edited" || msg.Author.ID != author.ID {
		t.Fatalf("got %q by %d, want the edited content by %d", msg.Content, msg.Author.ID, author.ID)
	}

	// A message the bot never saw isn't known.
	dispatch(&gateway.MessageDeleteEvent{ID: 2000, ChannelID: channelID, GuildID: 20})
	if _, ok := deleted.DeletedMessage(channelID, 2000); ok {
		t.Fatal("unknown message was kept")
	}

	// Bulk deletes keep the messages as well.
	dispatch(&gateway.MessageCreateEvent{Message: discord.Message{ID: 3000, ChannelID: channelID, Author: author, Content: "a"}})
	dispatch(&gateway.MessageCreateEvent{Message: discord.Message{ID: 3001, ChannelID: channelID, Author: author, Content: "b"}})
	dispatch(&gateway.MessageDeleteBulkEvent{IDs: []discord.MessageID{3000, 3001}, ChannelID: channelID})
	for _, id := range []discord.MessageID{3000, 3001} {
		if _, ok := deleted.DeletedMessage(channelID, id); !ok {
			t.Errorf("bulk deleted message %d wasn't kept", id)
		}
	}
}

func TestDeletedMessageStoreDropsOldest(t *testing.T) {
	s := newDeletedMessageStore(defaultstore.NewMessage(10))

	// One message per channel, so the inner store never drops one itself.
	total := maxDeletedMessages + 5
	for i := 1; i <= total; i++ {
		msg := discord.Message{ID: discord.MessageID(i), ChannelID: discord.ChannelID(i), Content: "x"}
		if err := s.MessageSet(&msg, false); err != nil {
			t.Fatal(err)
		}
		if err := s.MessageRemove(msg.ChannelID, msg.ID); err != nil {
			t.Fatal(err)
		}
	}

	if len(s.deleted) != maxDeletedMessages || len(s.order) != maxDeletedMessages {
		t.Fatalf("kept %d messages (%d keys), want %d", len(s.deleted), len(s.order), maxDeletedMessages)
	}
	for i := 1; i <= total; i++ {
		_, ok := s.DeletedMessage(discord.ChannelID(i), discord.MessageID(i))
		if want := i > total-maxDeletedMessages; ok != want {
			t.Errorf("message %d: kept = %v, want %v", i, ok, want)
		}
	}
}
