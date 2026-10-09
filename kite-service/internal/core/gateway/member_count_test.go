package gateway

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

func TestTrackMemberCounts(t *testing.T) {
	session := state.New("Bot token")
	trackMemberCounts(session)

	const guildID = discord.GuildID(1)
	count := func() uint64 {
		t.Helper()
		guild, err := session.Cabinet.Guild(guildID)
		if err != nil {
			t.Fatalf("guild not cached: %v", err)
		}
		return guild.ApproximateMembers
	}
	// Events go through the session like they do when Discord sends them, so
	// the cache is updated before the handlers run.
	dispatch := func(e any) { session.Session.Handler.Call(e) }

	dispatch(&gateway.GuildCreateEvent{
		Guild:       discord.Guild{ID: guildID, Name: "Kite HQ"},
		MemberCount: 10,
	})
	if got := count(); got != 10 {
		t.Fatalf("after create: got %d, want 10", got)
	}

	dispatch(&gateway.GuildMemberAddEvent{GuildID: guildID, Member: discord.Member{User: discord.User{ID: 5}}})
	dispatch(&gateway.GuildMemberAddEvent{GuildID: guildID, Member: discord.Member{User: discord.User{ID: 6}}})
	if got := count(); got != 12 {
		t.Fatalf("after joins: got %d, want 12", got)
	}

	dispatch(&gateway.GuildMemberRemoveEvent{GuildID: guildID, User: discord.User{ID: 5}})
	if got := count(); got != 11 {
		t.Fatalf("after leave: got %d, want 11", got)
	}

	// Server updates don't include the count and must not reset it.
	dispatch(&gateway.GuildUpdateEvent{Guild: discord.Guild{ID: guildID, Name: "Renamed"}})
	guild, _ := session.Cabinet.Guild(guildID)
	if guild.Name != "Renamed" || guild.ApproximateMembers != 11 {
		t.Fatalf("after update: got %q with %d members, want Renamed with 11", guild.Name, guild.ApproximateMembers)
	}

	// Events of servers that aren't cached are ignored.
	dispatch(&gateway.GuildMemberAddEvent{GuildID: 2, Member: discord.Member{User: discord.User{ID: 7}}})
	if _, err := session.Cabinet.Guild(2); err == nil {
		t.Fatal("uncached server was added")
	}
}
