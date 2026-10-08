package gateway

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store"
)

// memberCountStore keeps the member count of cached servers, which Discord
// only sends when the bot connects or joins a server. It's kept in
// ApproximateMembers of the server, so everything reading the cache sees it.
//
// The count is exact as long as the bot receives member join and leave events,
// which needs the server members intent. Without it the count is the one from
// when the bot connected.
type memberCountStore struct {
	store.GuildStore

	mu sync.Mutex
}

// GuildSet keeps the known count when the server is updated, as updates from
// Discord don't include it.
func (s *memberCountStore) GuildSet(guild *discord.Guild, update bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if guild.ApproximateMembers == 0 {
		if old, err := s.GuildStore.Guild(guild.ID); err == nil {
			cpy := *guild
			cpy.ApproximateMembers = old.ApproximateMembers
			guild = &cpy
		}
	}
	return s.GuildStore.GuildSet(guild, update)
}

// change applies fn to the member count of a cached server.
func (s *memberCountStore) change(guildID discord.GuildID, fn func(count uint64) uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	guild, err := s.GuildStore.Guild(guildID)
	if err != nil {
		return
	}
	guild.ApproximateMembers = fn(guild.ApproximateMembers)
	_ = s.GuildStore.GuildSet(guild, true)
}

// trackMemberCounts makes the cache of session keep member counts. It has to
// be called before any other handler is added, so they see the new count.
func trackMemberCounts(session *state.State) {
	counts := &memberCountStore{GuildStore: session.Cabinet.GuildStore}
	session.Cabinet.GuildStore = counts

	session.AddSyncHandler(func(e *gateway.GuildCreateEvent) {
		if e.Unavailable || e.MemberCount == 0 {
			return
		}
		counts.change(e.ID, func(uint64) uint64 { return e.MemberCount })
	})
	session.AddSyncHandler(func(e *gateway.GuildMemberAddEvent) {
		counts.change(e.GuildID, func(count uint64) uint64 { return count + 1 })
	})
	session.AddSyncHandler(func(e *gateway.GuildMemberRemoveEvent) {
		counts.change(e.GuildID, func(count uint64) uint64 {
			if count == 0 {
				return 0
			}
			return count - 1
		})
	})
}
