package guildevent

import (
	"slices"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func TestNewKnowsEveryEvent(t *testing.T) {
	for _, e := range []Event{&MessageDelete{}, &MemberUpdate{}} {
		got := New(e.EventType())
		if got == nil || got.EventType() != e.EventType() {
			t.Errorf("New(%q) = %T", e.EventType(), got)
		}
	}
	if New("GUILD_BAN_ADD") != nil {
		t.Error("New returned an event for a Discord event type")
	}
}

func TestMemberUpdateChanges(t *testing.T) {
	tests := map[string]struct {
		old  discord.Member
		new  gateway.GuildMemberUpdateEvent
		want []string
	}{
		"nickname set":      {discord.Member{}, gateway.GuildMemberUpdateEvent{Nick: "jack"}, []string{MemberChangeNickname}},
		"nickname removed":  {discord.Member{Nick: "jack"}, gateway.GuildMemberUpdateEvent{}, []string{MemberChangeNickname}},
		"server avatar set": {discord.Member{}, gateway.GuildMemberUpdateEvent{Avatar: "abc"}, []string{MemberChangeAvatar}},
		"screening passed":  {discord.Member{IsPending: true}, gateway.GuildMemberUpdateEvent{}, []string{MemberChangeScreeningPassed}},
		"still pending":     {discord.Member{IsPending: true}, gateway.GuildMemberUpdateEvent{IsPending: true}, nil},
		// Roles and timeouts don't run Member Update listeners.
		"roles changed": {discord.Member{Nick: "jack"}, gateway.GuildMemberUpdateEvent{Nick: "jack", RoleIDs: []discord.RoleID{1}}, nil},
		"several":       {discord.Member{IsPending: true}, gateway.GuildMemberUpdateEvent{Nick: "jack", Avatar: "abc"}, []string{MemberChangeNickname, MemberChangeAvatar, MemberChangeScreeningPassed}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := &MemberUpdate{GuildMemberUpdateEvent: tt.new, Old: tt.old}
			if got := e.Changes(); !slices.Equal(got, tt.want) {
				t.Errorf("Changes() = %v, want %v", got, tt.want)
			}
		})
	}
}
