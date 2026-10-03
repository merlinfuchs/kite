package discordevent

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

// MemberRemoveEvent is GUILD_MEMBER_REMOVE with what was known about the
// member before they left. Discord only sends the user, so the nickname and
// roles come from the member cache and are empty if the member wasn't cached.
type MemberRemoveEvent struct {
	*gateway.GuildMemberRemoveEvent

	Nick string `json:"nick,omitempty"`
	// RoleIDs are ordered from the highest role to the lowest.
	RoleIDs []discord.RoleID `json:"roles,omitempty"`
}

// Member returns the member as they were before leaving.
func (e *MemberRemoveEvent) Member() discord.Member {
	return discord.Member{
		User:    e.User,
		Nick:    e.Nick,
		RoleIDs: e.RoleIDs,
	}
}
