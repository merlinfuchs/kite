package eval

import (
"testing"

"github.com/diamondburned/arikawa/v3/discord"
"github.com/diamondburned/arikawa/v3/gateway"
"github.com/diamondburned/arikawa/v3/state"
"github.com/diamondburned/arikawa/v3/state/store"
)

// fakeDeletedMessageStore knows a fixed set of deleted messages.
type fakeDeletedMessageStore struct {
store.MessageStore
messages map[discord.MessageID]discord.Message
}

func (s fakeDeletedMessageStore) DeletedMessage(channelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, bool) {
msg, ok := s.messages[messageID]
if !ok || msg.ChannelID != channelID {
return nil, false
}
return &msg, true
}

// deletedMessageContext builds the context of a Message Delete event like
// NewContextFromEvent does, without the app user, which needs a Ready event.
func deletedMessageContext(id discord.MessageID, session *state.State) Context {
env := NewEventEnv(&gateway.MessageDeleteEvent{ID: id, ChannelID: 21, GuildID: testGuildID}, session)
return NewContext(Env{
"event":   env,
"user":    env.User,
"member":  env.Member,
"message": env.Message,
})
}

func TestMessageDeletePlaceholders(t *testing.T) {
session := testSession(t)
author := discord.User{ID: testUserID, Username: "jack"}
other := discord.User{ID: 42, Username: "someone"}
session.Cabinet.MessageStore = fakeDeletedMessageStore{
MessageStore: session.Cabinet.MessageStore,
messages: map[discord.MessageID]discord.Message{
// Sent in a server by a member that is cached.
1: {ID: 1, ChannelID: 21, GuildID: testGuildID, Author: author, Content: "deleted text"},
// Sent by someone who isn't cached as a member.
2: {ID: 2, ChannelID: 21, GuildID: testGuildID, Author: other, Content: "no member"},
},
}
// The state caches the member of the author when a message is created.
member := discord.Member{User: author, Nick: "Jacky", RoleIDs: []discord.RoleID{12}}
if err := session.Cabinet.MemberSet(testGuildID, &member, false); err != nil {
t.Fatal(err)
}

const template = "{{message.id}}|{{message.content}}|{{user.id}}|{{user.username}}"
tests := map[string]struct {
id   discord.MessageID
want string
}{
"with member":    {1, "1|deleted text|175928847299117063|jack"},
"without member": {2, "2|no member|42|someone"},
}
for name, tt := range tests {
c := deletedMessageContext(tt.id, session)
if got := evalString(t, c, template); got != tt.want {
t.Errorf("%s: got %q, want %q", name, got, tt.want)
}
}

// The bot never saw the message, so only its ID is known.
env := NewEventEnv(&gateway.MessageDeleteEvent{ID: 3, ChannelID: 21, GuildID: testGuildID}, session)
if env.User != nil {
t.Errorf("user is set for an unknown message: %#v", env.User)
}
if env.Message.ID != "3" || env.Message.Content != "" {
t.Errorf("unknown message: got ID %q and content %q", env.Message.ID, env.Message.Content)
}

// The nickname of the member is available, as for Message Create.
c := deletedMessageContext(1, session)
if got := evalString(t, c, "{{member.nick}}"); got != "Jacky" {
t.Fatalf("member nick: got %q, want %q", got, "Jacky")
}
}

func TestDeletedMessageWithoutStore(t *testing.T) {
// Sessions without a store that keeps deleted messages, and no session at
// all, fall back to the ID.
if _, ok := deletedMessage(nil, 1, 1); ok {
t.Fatal("found a message without a session")
}
if _, ok := deletedMessage(testSession(t), 1, 1); ok {
t.Fatal("found a message in the default store")
}
}