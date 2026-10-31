package gateway

import (
	"errors"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

type fakeCache struct {
	messages map[discord.MessageID]discord.Message
}

var errNotCached = errors.New("not cached")

func (c fakeCache) Message(channelID discord.ChannelID, id discord.MessageID) (*discord.Message, error) {
	m, ok := c.messages[id]
	if !ok || m.ChannelID != channelID {
		return nil, errNotCached
	}
	return &m, nil
}

func TestDeriveEvents(t *testing.T) {
	edited := discord.NewTimestamp(time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC))
	c := fakeCache{
		messages: map[discord.MessageID]discord.Message{
			7: {ID: 7, ChannelID: 5, GuildID: 2, Content: "hello"},
			9: {ID: 9, ChannelID: 5, GuildID: 2, Content: "hello, edited", EditedTimestamp: edited},
		},
	}

	tests := []struct {
		name  string
		event gateway.Event
		want  []ws.EventType
	}{
		{"cached message deleted", &gateway.MessageDeleteEvent{ID: 7, ChannelID: 5, GuildID: 2}, []ws.EventType{guildevent.MessageDeleteType}},
		// Listeners still run, with only the IDs as before.
		{"unknown message deleted", &gateway.MessageDeleteEvent{ID: 8, ChannelID: 5, GuildID: 2}, []ws.EventType{guildevent.MessageDeleteType}},
		// Message Delete listeners don't run for DMs and bulk deletes.
		{"message deleted in DM", &gateway.MessageDeleteEvent{ID: 7, ChannelID: 5}, nil},
		{"messages deleted in bulk", &gateway.MessageDeleteBulkEvent{IDs: []discord.MessageID{7}, ChannelID: 5, GuildID: 2}, nil},
		{"cached message edited", &gateway.MessageUpdateEvent{Message: discord.Message{ID: 7, ChannelID: 5, GuildID: 2, Content: "hello, edited", EditedTimestamp: edited}}, []ws.EventType{"MESSAGE_UPDATE"}},
		{"unknown message edited", &gateway.MessageUpdateEvent{Message: discord.Message{ID: 8, ChannelID: 5, GuildID: 2, EditedTimestamp: edited}}, []ws.EventType{"MESSAGE_UPDATE"}},
		// Link previews and pins update the message without editing it.
		{"link preview added", &gateway.MessageUpdateEvent{Message: discord.Message{ID: 7, ChannelID: 5, GuildID: 2, Content: "hello"}}, nil},
		{"cached link preview added after an edit", &gateway.MessageUpdateEvent{Message: discord.Message{ID: 9, ChannelID: 5, GuildID: 2, EditedTimestamp: edited}}, nil},
		{"message edited in DM", &gateway.MessageUpdateEvent{Message: discord.Message{ID: 7, ChannelID: 5, EditedTimestamp: edited}}, nil},
		{"unrelated event", &gateway.MessageCreateEvent{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveEvents(c, newEditTimes(), tt.event)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d events, want %d", len(got), len(tt.want))
			}
			for i, e := range got {
				if e == nil {
					t.Fatalf("event %d is nil", i)
				}
				if e.EventType() != tt.want[i] {
					t.Errorf("event %d is %q, want %q", i, e.EventType(), tt.want[i])
				}
			}
		})
	}
}

func TestTrackDerivedEventsKeepsDeletedMessage(t *testing.T) {
	session := state.New("Bot token")
	derived := make(chan gateway.Event, 1)
	trackDerivedEvents(session, func(e gateway.Event) { derived <- e })

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
	dispatch(&gateway.MessageDeleteEvent{ID: 1000, ChannelID: channelID, GuildID: 20})

	var e gateway.Event
	select {
	case e = <-derived:
	case <-time.After(time.Second):
		t.Fatal("no event derived from the delete")
	}

	if _, err := session.Cabinet.Message(channelID, 1000); err == nil {
		t.Fatal("deleted message is still in the cache")
	}

	del, ok := e.(*guildevent.MessageDelete)
	if !ok {
		t.Fatalf("got %T, want *guildevent.MessageDelete", e)
	}
	if del.ID != 1000 || del.ChannelID != channelID || del.GuildID != 20 {
		t.Errorf("IDs = %d, %d, %d, want those of the deleted message", del.ID, del.ChannelID, del.GuildID)
	}
	if del.Message == nil || del.Message.Content != "hello, edited" || del.Message.Author.ID != author.ID {
		t.Fatalf("message = %+v, want the edited content by %d", del.Message, author.ID)
	}
}

func TestTrackDerivedEventsOnlyEdits(t *testing.T) {
	session := state.New("Bot token")
	derived := make(chan gateway.Event, 3)
	trackDerivedEvents(session, func(e gateway.Event) { derived <- e })
	dispatch := func(e any) { session.Session.Handler.Call(e) }

	msg := discord.Message{ID: 1000, ChannelID: 10, GuildID: 20, Author: discord.User{ID: 5}, Content: "hello"}
	dispatch(&gateway.MessageCreateEvent{Message: msg})

	// The link preview of the message, then an edit and the link preview of
	// the edited content, which keeps the time of the edit.
	dispatch(&gateway.MessageUpdateEvent{Message: msg})
	msg.Content = "hello, edited"
	msg.EditedTimestamp = discord.NewTimestamp(time.Now())
	dispatch(&gateway.MessageUpdateEvent{Message: msg})
	dispatch(&gateway.MessageUpdateEvent{Message: msg})

	select {
	case e := <-derived:
		edit, ok := e.(*guildevent.MessageEdit)
		if !ok || edit.Content != "hello, edited" {
			t.Fatalf("got %+v, want the edit", e)
		}
	case <-time.After(time.Second):
		t.Fatal("no event derived from the edit")
	}
	select {
	case e := <-derived:
		t.Fatalf("got %+v, want only the edit", e)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestDeriveEventsRemembersEditsOfUncachedMessages(t *testing.T) {
	edits := newEditTimes()
	update := func(edited time.Time) []guildevent.Event {
		return deriveEvents(fakeCache{}, edits, &gateway.MessageUpdateEvent{Message: discord.Message{
			ID: 8, ChannelID: 5, GuildID: 2, EditedTimestamp: discord.NewTimestamp(edited),
		}})
	}

	first := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	if len(update(first)) != 1 {
		t.Fatal("edit of an uncached message isn't an edit")
	}
	// The link preview of the edited content keeps the time of the edit.
	if len(update(first)) != 0 {
		t.Fatal("link preview after the edit is an edit")
	}
	if len(update(first.Add(time.Minute))) != 1 {
		t.Fatal("second edit isn't an edit")
	}
}
