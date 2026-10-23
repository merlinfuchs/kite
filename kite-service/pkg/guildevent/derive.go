package guildevent

import (
	"slices"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

// MaxRoleEvents bounds the role events of a single member update. Each runs
// the flows of the listeners, so taking all roles of a member at once
// shouldn't run them hundreds of times.
const MaxRoleEvents = 25

// FromMemberUpdate derives the role and timeout events of a member update.
// Discord only sends the member after the change, so before is the member as
// it was cached, nil if it wasn't.
//
// Without the member from before, its old roles are unknown and no role
// events are derived. A timeout that's active now is still reported as new,
// as that's by far the most likely reason for the update.
func FromMemberUpdate(before *discord.Member, e *gateway.GuildMemberUpdateEvent, now time.Time) []Event {
	var after discord.Member
	if before != nil {
		after = *before
		after.RoleIDs = slices.Clone(before.RoleIDs)
	}
	e.UpdateMember(&after)

	var res []Event

	if before != nil {
		for _, id := range after.RoleIDs {
			if !slices.Contains(before.RoleIDs, id) {
				res = append(res, &MemberRoleAdd{MemberRole{GuildID: e.GuildID, Member: after, RoleID: id}})
			}
		}
		for _, id := range before.RoleIDs {
			if !slices.Contains(after.RoleIDs, id) {
				res = append(res, &MemberRoleRemove{MemberRole{GuildID: e.GuildID, Member: after, RoleID: id}})
			}
		}
		if len(res) > MaxRoleEvents {
			res = res[:MaxRoleEvents]
		}
	}

	var wasTimedOut bool
	if before != nil {
		wasTimedOut = timeoutActive(before.CommunicationDisabledUntil, now)
	}
	isTimedOut := timeoutActive(after.CommunicationDisabledUntil, now)

	timeout := MemberTimeout{GuildID: e.GuildID, Member: after}
	switch {
	case !wasTimedOut && isTimedOut:
		res = append(res, &MemberTimeoutAdd{timeout})
	case wasTimedOut && isTimedOut && !before.CommunicationDisabledUntil.Time().Equal(after.CommunicationDisabledUntil.Time()):
		// A moderator timing out a member who already is sets a new end.
		res = append(res, &MemberTimeoutAdd{timeout})
	case wasTimedOut && !isTimedOut:
		res = append(res, &MemberTimeoutRemove{timeout})
	}

	return res
}

func timeoutActive(until discord.Timestamp, now time.Time) bool {
	return until.IsValid() && until.Time().After(now)
}

// FromChannelUpdate adds the channel as it was cached, nil if it wasn't. It
// returns nil for channels outside of a server, and if nothing it reports on
// changed, which filters out updates Discord sends when only the position
// changed.
func FromChannelUpdate(before *discord.Channel, e *gateway.ChannelUpdateEvent) *ChannelUpdate {
	if e.GuildID == 0 {
		return nil
	}

	res := &ChannelUpdate{Channel: e.Channel}
	if before != nil {
		res.Before = before
		res.Changes = ChannelChanges(*before, e.Channel)
		if len(res.Changes) == 0 {
			return nil
		}
	}
	return res
}

// ChannelChanges names the settings that differ between two versions of a
// channel, in the order they're shown in Discord. The position is left out:
// dragging one channel makes Discord send an update for every channel that
// moved along with it.
func ChannelChanges(before, after discord.Channel) []string {
	var res []string
	if before.Name != after.Name {
		res = append(res, "name")
	}
	if before.Type != after.Type {
		res = append(res, "type")
	}
	if before.Topic != after.Topic {
		res = append(res, "topic")
	}
	if before.NSFW != after.NSFW {
		res = append(res, "nsfw")
	}
	if before.UserRateLimit != after.UserRateLimit {
		res = append(res, "slowmode")
	}
	if before.ParentID != after.ParentID {
		res = append(res, "category")
	}
	if !overwritesEqual(before.Overwrites, after.Overwrites) {
		res = append(res, "permissions")
	}
	if before.VoiceBitrate != after.VoiceBitrate {
		res = append(res, "bitrate")
	}
	if before.VoiceUserLimit != after.VoiceUserLimit {
		res = append(res, "user limit")
	}
	if before.RTCRegionID != after.RTCRegionID {
		res = append(res, "region")
	}
	return res
}

// overwritesEqual ignores the order, which Discord doesn't keep stable.
func overwritesEqual(a, b []discord.Overwrite) bool {
	if len(a) != len(b) {
		return false
	}
	for _, o := range a {
		if !slices.Contains(b, o) {
			return false
		}
	}
	return true
}
