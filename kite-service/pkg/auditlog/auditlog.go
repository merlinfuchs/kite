// Package auditlog describes the entries of a server's audit log for Audit Log
// Entry event listeners: the name of their action, which the Event Filter
// matches, and what they changed in words.
package auditlog

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

// Timeouts are member updates in the audit log, so Action names them
// separately.
const (
	ActionMemberTimeoutAdd    = "member_timeout_add"
	ActionMemberTimeoutRemove = "member_timeout_remove"
)

const keyTimeout = "communication_disabled_until"

// actionNames are Discord's names of the actions, lowercased.
var actionNames = map[discord.AuditLogEvent]string{
	1:   "guild_update",
	10:  "channel_create",
	11:  "channel_update",
	12:  "channel_delete",
	13:  "channel_overwrite_create",
	14:  "channel_overwrite_update",
	15:  "channel_overwrite_delete",
	20:  "member_kick",
	21:  "member_prune",
	22:  "member_ban_add",
	23:  "member_ban_remove",
	24:  "member_update",
	25:  "member_role_update",
	26:  "member_move",
	27:  "member_disconnect",
	28:  "bot_add",
	30:  "role_create",
	31:  "role_update",
	32:  "role_delete",
	40:  "invite_create",
	41:  "invite_update",
	42:  "invite_delete",
	50:  "webhook_create",
	51:  "webhook_update",
	52:  "webhook_delete",
	60:  "emoji_create",
	61:  "emoji_update",
	62:  "emoji_delete",
	72:  "message_delete",
	73:  "message_bulk_delete",
	74:  "message_pin",
	75:  "message_unpin",
	80:  "integration_create",
	81:  "integration_update",
	82:  "integration_delete",
	83:  "stage_instance_create",
	84:  "stage_instance_update",
	85:  "stage_instance_delete",
	90:  "sticker_create",
	91:  "sticker_update",
	92:  "sticker_delete",
	100: "guild_scheduled_event_create",
	101: "guild_scheduled_event_update",
	102: "guild_scheduled_event_delete",
	110: "thread_create",
	111: "thread_update",
	112: "thread_delete",
	121: "application_command_permission_update",
	130: "soundboard_sound_create",
	131: "soundboard_sound_update",
	132: "soundboard_sound_delete",
	140: "auto_moderation_rule_create",
	141: "auto_moderation_rule_update",
	142: "auto_moderation_rule_delete",
	143: "auto_moderation_block_message",
	144: "auto_moderation_flag_to_channel",
	145: "auto_moderation_user_communication_disabled",
	146: "auto_moderation_quarantine_user",
	150: "creator_monetization_request_created",
	151: "creator_monetization_terms_accepted",
	163: "onboarding_prompt_create",
	164: "onboarding_prompt_update",
	165: "onboarding_prompt_delete",
	166: "onboarding_create",
	167: "onboarding_update",
	190: "home_settings_create",
	191: "home_settings_update",
	192: "voice_channel_status_create",
	193: "voice_channel_status_delete",
}

// Action names the action of an entry, like member_ban_add. A member update
// that times a member out or changes when their timeout ends is
// member_timeout_add, one that removes it member_timeout_remove. Actions Kite
// doesn't know yet are named by their number.
func Action(e discord.AuditLogEntry) string {
	if e.ActionType == discord.MemberUpdate {
		if c, ok := change(e, keyTimeout); ok {
			if isNull(c.NewValue) {
				return ActionMemberTimeoutRemove
			}
			return ActionMemberTimeoutAdd
		}
	}
	if name, ok := actionNames[e.ActionType]; ok {
		return name
	}
	return fmt.Sprint(e.ActionType)
}

// TargetKind is what the target of an entry is.
type TargetKind int

const (
	TargetOther TargetKind = iota
	TargetUser
	TargetRole
	TargetChannel
)

// Target returns what the target of an entry is. Messages are targeted by
// their author.
func Target(e discord.AuditLogEntry) TargetKind {
	switch e.ActionType {
	case 20, 22, 23, 24, 25, 28, 72, 74, 75, 143, 144, 145, 146:
		return TargetUser
	case 30, 31, 32:
		return TargetRole
	case 10, 11, 12, 13, 14, 15, 73, 110, 111, 112, 192, 193:
		return TargetChannel
	}
	return TargetOther
}

// ChannelID is the channel an entry is about, either because the channel is
// its target or because it happened in it, like a deleted message. Zero if
// there's none.
func ChannelID(e discord.AuditLogEntry) discord.ChannelID {
	if e.Options.ChannelID.IsValid() {
		return e.Options.ChannelID
	}
	if Target(e) == TargetChannel {
		return discord.ChannelID(e.TargetID)
	}
	return 0
}

// RolesAdded are the roles a member role update added, with their names as
// they were then.
func RolesAdded(e discord.AuditLogEntry) []discord.Role {
	return roles(e, "$add")
}

// RolesRemoved are the roles a member role update removed.
func RolesRemoved(e discord.AuditLogEntry) []discord.Role {
	return roles(e, "$remove")
}

func roles(e discord.AuditLogEntry, key string) []discord.Role {
	c, ok := change(e, key)
	if !ok {
		return nil
	}
	var res []discord.Role
	if err := json.Unmarshal(c.NewValue, &res); err != nil {
		return nil
	}
	return res
}

// RoleNames lists the names of roles, separated by commas.
func RoleNames(roles []discord.Role) string {
	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name
	}
	return strings.Join(names, ", ")
}

// TimeoutUntil is when the timeout of a member_timeout_add ends, zero for
// other entries.
func TimeoutUntil(e discord.AuditLogEntry) time.Time {
	c, ok := change(e, keyTimeout)
	if !ok {
		return time.Time{}
	}
	var until time.Time
	if err := json.Unmarshal(c.NewValue, &until); err != nil {
		return time.Time{}
	}
	return until
}

// Changes describes what an entry changed, one line per setting, like
// "name: general → chat". Discord's names of the settings are kept, as they
// can't all be put into words, and times are shown as Discord timestamps.
func Changes(e discord.AuditLogEntry) string {
	lines := make([]string, 0, len(e.Changes))
	for _, c := range e.Changes {
		switch c.Key {
		case "$add", "$remove":
			label := "roles added"
			if c.Key == "$remove" {
				label = "roles removed"
			}
			lines = append(lines, label+": "+RoleNames(roles(e, string(c.Key))))
		default:
			lines = append(lines, fmt.Sprintf("%s: %s → %s", c.Key, describeValue(c.OldValue), describeValue(c.NewValue)))
		}
	}
	return strings.Join(lines, "\n")
}

func describeValue(raw []byte) string {
	if isNull(raw) {
		return "none"
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return fmt.Sprintf("<t:%d:f>", t.Unix())
		}
		if s == "" {
			return "none"
		}
		return s
	}
	return string(raw)
}

func change(e discord.AuditLogEntry, key string) (discord.AuditLogChange, bool) {
	for _, c := range e.Changes {
		if string(c.Key) == key {
			return c, true
		}
	}
	return discord.AuditLogChange{}, false
}

func isNull(raw []byte) bool {
	s := strings.TrimSpace(string(raw))
	return s == "" || s == "null"
}
