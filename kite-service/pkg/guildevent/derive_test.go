package guildevent

import (
	"slices"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func memberUpdate(roles []discord.RoleID, until time.Time) *gateway.GuildMemberUpdateEvent {
	e := &gateway.GuildMemberUpdateEvent{GuildID: 2, User: discord.User{ID: 1}, RoleIDs: roles}
	if !until.IsZero() {
		e.CommunicationDisabledUntil = discord.NewTimestamp(until)
	}
	return e
}

func member(roles []discord.RoleID, until time.Time) *discord.Member {
	m := &discord.Member{User: discord.User{ID: 1}, RoleIDs: roles}
	if !until.IsZero() {
		m.CommunicationDisabledUntil = discord.NewTimestamp(until)
	}
	return m
}

func types(events []Event) []string {
	var res []string
	for _, e := range events {
		res = append(res, string(e.EventType()))
	}
	return res
}

const (
	roleAdd       = string(MemberRoleAddType)
	roleRemove    = string(MemberRoleRemoveType)
	timeoutAdd    = string(MemberTimeoutAddType)
	timeoutRemove = string(MemberTimeoutRemoveType)
)

func TestFromMemberUpdate(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)
	muchLater := now.Add(2 * time.Hour)
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name   string
		before *discord.Member
		event  *gateway.GuildMemberUpdateEvent
		want   []string
	}{
		{"role added", member([]discord.RoleID{10}, time.Time{}), memberUpdate([]discord.RoleID{10, 11}, time.Time{}), []string{roleAdd}},
		{"role removed", member([]discord.RoleID{10, 11}, time.Time{}), memberUpdate([]discord.RoleID{11}, time.Time{}), []string{roleRemove}},
		{"roles swapped", member([]discord.RoleID{10}, time.Time{}), memberUpdate([]discord.RoleID{11}, time.Time{}), []string{roleAdd, roleRemove}},
		{"nickname only", member([]discord.RoleID{10}, time.Time{}), memberUpdate([]discord.RoleID{10}, time.Time{}), nil},
		{"timed out", member(nil, time.Time{}), memberUpdate(nil, later), []string{timeoutAdd}},
		{"timeout removed", member(nil, later), memberUpdate(nil, time.Time{}), []string{timeoutRemove}},
		{"timeout changed", member(nil, later), memberUpdate(nil, muchLater), []string{timeoutAdd}},
		// Discord sends nothing when a timeout runs out, so a later update
		// with the old timestamp must not look like a new timeout or removal.
		{"expired timeout", member(nil, earlier), memberUpdate(nil, earlier), nil},
		{"expired timeout cleared", member(nil, earlier), memberUpdate(nil, time.Time{}), nil},
		{"uncached member timed out", nil, memberUpdate([]discord.RoleID{10}, later), []string{timeoutAdd}},
		{"uncached member", nil, memberUpdate([]discord.RoleID{10}, time.Time{}), nil},
		{"role and timeout at once", member(nil, time.Time{}), memberUpdate([]discord.RoleID{10}, later), []string{roleAdd, timeoutAdd}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types(FromMemberUpdate(tt.before, tt.event, now))
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromMemberUpdateDetails(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)
	before := member([]discord.RoleID{10}, time.Time{})
	before.Nick = "old"
	before.Joined = discord.NewTimestamp(now.Add(-24 * time.Hour))

	event := memberUpdate([]discord.RoleID{10, 11}, later)
	event.Nick = "new"

	got := FromMemberUpdate(before, event, now)
	if len(got) != 2 {
		t.Fatalf("got %d events, want 2", len(got))
	}

	role := got[0].(*MemberRoleAdd)
	if role.RoleID != 11 || role.GuildID != 2 {
		t.Errorf("role event = %+v", role)
	}
	// The member is the one after the change, with what the update doesn't
	// carry kept from the cache.
	if role.Member.Nick != "new" || !role.Member.Joined.IsValid() || len(role.Member.RoleIDs) != 2 {
		t.Errorf("member = %+v", role.Member)
	}
	if len(before.RoleIDs) != 1 {
		t.Error("the cached member was modified")
	}

	timeout := got[1].(*MemberTimeoutAdd)
	if timeout.Member.CommunicationDisabledUntil.Time().Unix() != later.Unix() {
		t.Errorf("until = %v, want %v", timeout.Member.CommunicationDisabledUntil.Time(), later)
	}
}

func TestFromMemberUpdateLimitsRoleEvents(t *testing.T) {
	var roles []discord.RoleID
	for i := range MaxRoleEvents + 10 {
		roles = append(roles, discord.RoleID(100+i))
	}

	got := FromMemberUpdate(member(roles, time.Time{}), memberUpdate(nil, time.Time{}), time.Now())
	if len(got) != MaxRoleEvents {
		t.Errorf("got %d events, want %d", len(got), MaxRoleEvents)
	}
}

func TestFromChannelUpdate(t *testing.T) {
	before := &discord.Channel{ID: 1, GuildID: 2, Name: "general", Position: 1}

	renamed := FromChannelUpdate(before, &gateway.ChannelUpdateEvent{Channel: discord.Channel{ID: 1, GuildID: 2, Name: "chat", Position: 1}})
	if renamed == nil || renamed.Before != before || !slices.Equal(renamed.Changes, []string{"name"}) {
		t.Errorf("renamed = %+v", renamed)
	}
	if moved := FromChannelUpdate(before, &gateway.ChannelUpdateEvent{Channel: discord.Channel{ID: 1, GuildID: 2, Name: "general", Position: 4}}); moved != nil {
		t.Errorf("moved = %+v, want nil", moved)
	}
	// Without the channel from before, every update goes through.
	if unknown := FromChannelUpdate(nil, &gateway.ChannelUpdateEvent{Channel: discord.Channel{ID: 1, GuildID: 2}}); unknown == nil || unknown.Before != nil {
		t.Errorf("unknown = %+v", unknown)
	}
	if dm := FromChannelUpdate(nil, &gateway.ChannelUpdateEvent{Channel: discord.Channel{ID: 1}}); dm != nil {
		t.Errorf("dm = %+v, want nil", dm)
	}
}

func TestChannelChanges(t *testing.T) {
	before := discord.Channel{
		ID:         1,
		Name:       "general",
		Position:   1,
		Overwrites: []discord.Overwrite{{ID: 1, Allow: 1}, {ID: 2, Deny: 2}},
	}

	tests := []struct {
		name  string
		after func(c *discord.Channel)
		want  []string
	}{
		{"nothing", func(c *discord.Channel) {}, nil},
		{"position only", func(c *discord.Channel) { c.Position = 5 }, nil},
		{"overwrites reordered", func(c *discord.Channel) {
			c.Overwrites = []discord.Overwrite{{ID: 2, Deny: 2}, {ID: 1, Allow: 1}}
		}, nil},
		{"permissions", func(c *discord.Channel) {
			c.Overwrites = []discord.Overwrite{{ID: 1, Allow: 3}, {ID: 2, Deny: 2}}
		}, []string{"permissions"}},
		{"name and slowmode", func(c *discord.Channel) {
			c.Name = "chat"
			c.UserRateLimit = 10
		}, []string{"name", "slowmode"}},
		{"category", func(c *discord.Channel) { c.ParentID = 9 }, []string{"category"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			after := before
			after.Overwrites = slices.Clone(before.Overwrites)
			tt.after(&after)
			if got := ChannelChanges(before, after); !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKnowsEveryEvent(t *testing.T) {
	for _, e := range []Event{&MemberRoleAdd{}, &MemberRoleRemove{}, &MemberTimeoutAdd{}, &MemberTimeoutRemove{}, &ChannelUpdate{}} {
		got := New(e.EventType())
		if got == nil || got.EventType() != e.EventType() {
			t.Errorf("New(%q) = %T", e.EventType(), got)
		}
	}
	if New("GUILD_BAN_ADD") != nil {
		t.Error("New returned an event for a Discord event type")
	}
}
