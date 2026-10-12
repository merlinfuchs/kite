package gateway

import (
	"errors"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

type fakeCache struct {
	members  map[discord.UserID]discord.Member
	channels map[discord.ChannelID]discord.Channel
}

var errNotCached = errors.New("not cached")

func (c fakeCache) Member(_ discord.GuildID, id discord.UserID) (*discord.Member, error) {
	m, ok := c.members[id]
	if !ok {
		return nil, errNotCached
	}
	return &m, nil
}

func (c fakeCache) Channel(id discord.ChannelID) (*discord.Channel, error) {
	ch, ok := c.channels[id]
	if !ok {
		return nil, errNotCached
	}
	return &ch, nil
}

func TestDeriveEvents(t *testing.T) {
	now := time.Now()
	c := fakeCache{
		members: map[discord.UserID]discord.Member{
			1: {User: discord.User{ID: 1}, RoleIDs: []discord.RoleID{10}},
		},
		channels: map[discord.ChannelID]discord.Channel{
			5: {ID: 5, GuildID: 2, Name: "general"},
		},
	}

	tests := []struct {
		name  string
		event gateway.Event
		want  []ws.EventType
	}{
		{
			"role added to cached member",
			&gateway.GuildMemberUpdateEvent{GuildID: 2, User: discord.User{ID: 1}, RoleIDs: []discord.RoleID{10, 11}},
			[]ws.EventType{guildevent.MemberRoleAddType},
		},
		{
			// The old roles are unknown, so nothing can be said about them.
			"roles of uncached member",
			&gateway.GuildMemberUpdateEvent{GuildID: 2, User: discord.User{ID: 3}, RoleIDs: []discord.RoleID{10}},
			nil,
		},
		{"channel renamed", &gateway.ChannelUpdateEvent{Channel: discord.Channel{ID: 5, GuildID: 2, Name: "chat"}}, []ws.EventType{guildevent.ChannelUpdateType}},
		{"channel unchanged", &gateway.ChannelUpdateEvent{Channel: discord.Channel{ID: 5, GuildID: 2, Name: "general"}}, nil},
		// Discord's own events go to listeners as they are.
		{"ban", &gateway.GuildBanAddEvent{GuildID: 2, User: discord.User{ID: 1}}, nil},
		{"channel created", &gateway.ChannelCreateEvent{Channel: discord.Channel{ID: 6, GuildID: 2}}, nil},
		{"unrelated event", &gateway.MessageCreateEvent{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveEvents(c, tt.event, now)
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

func TestDeriveEventsKeepsChannelBefore(t *testing.T) {
	c := fakeCache{channels: map[discord.ChannelID]discord.Channel{
		5: {ID: 5, GuildID: 2, Name: "general"},
	}}

	got := deriveEvents(c, &gateway.ChannelUpdateEvent{Channel: discord.Channel{ID: 5, GuildID: 2, Name: "chat"}}, time.Now())
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1", len(got))
	}
	update := got[0].(*guildevent.ChannelUpdate)
	if update.Before == nil || update.Before.Name != "general" {
		t.Errorf("before = %+v, want the cached channel", update.Before)
	}
	if len(update.Changes) != 1 || update.Changes[0] != "name" {
		t.Errorf("changes = %v, want [name]", update.Changes)
	}
}
