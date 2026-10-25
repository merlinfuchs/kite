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
	members  map[discord.UserID]discord.Member
	roles    map[discord.RoleID]discord.Role
}

var errNotCached = errors.New("not cached")

func (c fakeCache) Message(channelID discord.ChannelID, id discord.MessageID) (*discord.Message, error) {
	m, ok := c.messages[id]
	if !ok || m.ChannelID != channelID {
		return nil, errNotCached
	}
	return &m, nil
}

func (c fakeCache) Member(guildID discord.GuildID, id discord.UserID) (*discord.Member, error) {
	m, ok := c.members[id]
	if !ok {
		return nil, errNotCached
	}
	return &m, nil
}

func (c fakeCache) Role(guildID discord.GuildID, id discord.RoleID) (*discord.Role, error) {
	r, ok := c.roles[id]
	if !ok {
		return nil, errNotCached
	}
	return &r, nil
}

const (
	boosterRole = discord.RoleID(30)
	botRole     = discord.RoleID(31)
	plainRole   = discord.RoleID(32)
)

func TestDeriveEvents(t *testing.T) {
	c := fakeCache{
		messages: map[discord.MessageID]discord.Message{
			7: {ID: 7, ChannelID: 5, GuildID: 2, Content: "hello"},
		},
		members: map[discord.UserID]discord.Member{
			40: {User: discord.User{ID: 40}, RoleIDs: []discord.RoleID{boosterRole, plainRole}},
			41: {User: discord.User{ID: 41}, RoleIDs: []discord.RoleID{botRole, plainRole}},
		},
		roles: map[discord.RoleID]discord.Role{
			boosterRole: {ID: boosterRole, Managed: true},
			botRole:     {ID: botRole, Managed: true, Tags: discord.RoleTags{BotID: 99}},
			plainRole:   {ID: plainRole},
		},
	}

	memberUpdate := func(userID discord.UserID, roles ...discord.RoleID) *gateway.GuildMemberUpdateEvent {
		return &gateway.GuildMemberUpdateEvent{GuildID: 2, User: discord.User{ID: userID}, RoleIDs: roles}
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
		{"unrelated event", &gateway.MessageCreateEvent{}, nil},
		{"booster role taken away", memberUpdate(40, plainRole), []ws.EventType{guildevent.BoostRemoveType}},
		{"booster role kept", memberUpdate(40, boosterRole), nil},
		{"other role taken away", memberUpdate(40, boosterRole, plainRole), nil},
		// Bot and integration roles are managed too, but aren't the booster role.
		{"bot role taken away", memberUpdate(41, plainRole), nil},
		// Without the cached member there's nothing to compare with.
		{"uncached member", memberUpdate(42), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveEvents(c, tt.event)
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

func TestTrackDerivedEventsBoostRemove(t *testing.T) {
	session := state.New("Bot token")
	derived := make(chan gateway.Event, 1)
	trackDerivedEvents(session, func(e gateway.Event) { derived <- e })

	dispatch := func(e any) { session.Session.Handler.Call(e) }

	const guildID = discord.GuildID(20)
	user := discord.User{ID: 5, Username: "jack"}

	dispatch(&gateway.GuildCreateEvent{Guild: discord.Guild{
		ID: guildID,
		Roles: []discord.Role{
			{ID: discord.RoleID(guildID), Name: "@everyone"},
			{ID: boosterRole, Name: "Server Booster", Managed: true},
		},
	}})
	dispatch(&gateway.GuildMemberAddEvent{GuildID: guildID, Member: discord.Member{
		User:    user,
		RoleIDs: []discord.RoleID{boosterRole},
	}})
	dispatch(&gateway.GuildMemberUpdateEvent{GuildID: guildID, User: user})

	var e gateway.Event
	select {
	case e = <-derived:
	case <-time.After(time.Second):
		t.Fatal("no event derived from the booster role being taken away")
	}

	removed, ok := e.(*guildevent.BoostRemove)
	if !ok {
		t.Fatalf("got %T, want *guildevent.BoostRemove", e)
	}
	if removed.GuildID != guildID || removed.User.ID != user.ID {
		t.Errorf("got %d in %d, want %d in %d", removed.User.ID, removed.GuildID, user.ID, guildID)
	}

	// The cache has the member without the role now, so another update
	// derives nothing.
	dispatch(&gateway.GuildMemberUpdateEvent{GuildID: guildID, User: user})
	select {
	case e := <-derived:
		t.Fatalf("derived %T from a member that wasn't boosting", e)
	case <-time.After(100 * time.Millisecond):
	}
}
